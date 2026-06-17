package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"ai-content-platform/database"
	"ai-content-platform/middleware"
	"ai-content-platform/models"
	"ai-content-platform/services"

	"github.com/gin-gonic/gin"
)

var aiService = services.NewAIService()

// CreateContent creates a new content
func CreateContent(c *gin.Context) {
	userID, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input struct {
		Title            string                `json:"title" binding:"required"`
		Summary          string                `json:"summary"`
		Type             models.ContentType    `json:"type" binding:"required"`
		InputType        models.ContentInputType `json:"input_type" binding:"required"`
		ContentData      string                `json:"content_data"`
		CoverImage       string                `json:"cover_image"`
		Tags             string                `json:"tags"`
		SourceURL        string                `json:"source_url"`
		Status           string                `json:"status"`
		Blocks           []models.ContentBlock `json:"blocks"`
		Category         string                `json:"category"`
		MetaTags         string                `json:"meta_tags"`
		GeneratedContent string                `json:"generated_content"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category := input.Category
	if category == "" {
		if cat, err := aiService.DetectCategory(input.Title + " " + input.ContentData); err == nil {
			category = cat
		}
	}

	metaTags := input.MetaTags
	if metaTags == "" {
		if tags, err := aiService.GenerateMetaTags(input.Title + " " + input.ContentData); err == nil {
			metaTags = tags
		}
	}

	content := models.Content{
		UserID:           userID,
		Title:            input.Title,
		Summary:          input.Summary,
		Type:             input.Type,
		InputType:        input.InputType,
		ContentData:      input.ContentData,
		CoverImage:       input.CoverImage,
		Tags:             input.Tags,
		SourceURL:        input.SourceURL,
		Status:           input.Status,
		Category:         category,
		MetaTags:         metaTags,
		GeneratedContent: input.GeneratedContent,
	}

	// Set default status to draft if not provided
	if content.Status == "" {
		content.Status = "draft"
	}

	// Check if using memory store
	if database.UseMemory {
		err := database.MemoryStore.CreateContent(&content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create content"})
			return
		}

		// Create content blocks if provided
		for i := range input.Blocks {
			input.Blocks[i].ContentID = content.ID
			if err := database.MemoryStore.CreateContentBlock(&input.Blocks[i]); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create content block"})
				return
			}
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Content created successfully",
			"content": content,
		})
		return
	}

	if err := database.DB.Create(&content).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create content"})
		return
	}

	// Create content blocks if provided
	for _, block := range input.Blocks {
		block.ContentID = content.ID
		if err := database.DB.Create(&block).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create content block"})
			return
		}
	}

	// Fetch the created content with associated blocks
	if err := database.DB.Preload("Blocks").First(&content, content.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch created content"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Content created successfully",
		"content": content,
	})
}

// GetContent retrieves a specific content by ID
func GetContent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content ID"})
		return
	}

	var content models.Content

	if database.UseMemory {
		contentPtr, err := database.MemoryStore.GetContentByID(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Content not found"})
			return
		}
		content = *contentPtr

		// Enrich with user info
		if user, err := database.MemoryStore.GetUserByID(content.UserID); err == nil {
			content.User = *user
		}
	} else {
		if err := database.DB.Preload("User").Preload("Blocks").Preload("Distributors").First(&content, uint(id)).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Content not found"})
			return
		}
	}

	// Check if user has permission to access this content
	_, exists := middleware.GetUserFromContext(c)
	if exists && !middleware.HasPermission(c, content.UserID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"content": content,
	})
}

// GetContents retrieves all contents with optional filters
func GetContents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	title := c.Query("title")
	contentType := c.Query("type")
	status := c.Query("status")
	category := c.Query("category")

	// Check if using memory store
	if database.UseMemory {
		userID, _ := middleware.GetUserFromContext(c)
		role, _ := middleware.GetRoleFromContext(c)
		isAdmin := role == "admin"

		contents := database.MemoryStore.GetContents(userID, isAdmin)

		// Apply filters
		var filtered []models.Content
		for _, content := range contents {
			if title != "" && !strings.Contains(strings.ToLower(content.Title), strings.ToLower(title)) {
				continue
			}
			if contentType != "" && string(content.Type) != contentType {
				continue
			}
			if status != "" && content.Status != status {
				continue
			}
			if category != "" && content.Category != category {
				continue
			}
			filtered = append(filtered, content)
		}

		// Apply pagination
		total := len(filtered)
		start := (page - 1) * limit
		end := start + limit
		if end > total {
			end = total
		}
		if start > total {
			start = total
		}

		// Enrich with user info
		for i := start; i < end; i++ {
			if user, err := database.MemoryStore.GetUserByID(filtered[i].UserID); err == nil {
				filtered[i].User = *user
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"contents": filtered[start:end],
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		})
		return
	}

	offset := (page - 1) * limit

	query := database.DB.Preload("User").Preload("Blocks").Preload("Distributors")

	// Apply filters
	if title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}
	if contentType != "" {
		query = query.Where("type = ?", contentType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// Check if user is requesting their own content only
	userID, exists := middleware.GetUserFromContext(c)
	if exists {
		role, _ := middleware.GetRoleFromContext(c)
		if role != "admin" {
			query = query.Where("user_id = ?", userID)
		}
	}

	var contents []models.Content
	var total int64

	if err := query.Offset(offset).Limit(limit).Find(&contents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch contents"})
		return
	}

	if err := query.Model(&models.Content{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count contents"})
		return
	}

	// Enrich with user info for in-memory style display
	for i := range contents {
		var user models.User
		if err := database.DB.Where("id = ?", contents[i].UserID).First(&user).Error; err == nil {
			user.Password = ""
			contents[i].User = user
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"contents": contents,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// UpdateContent updates an existing content
func UpdateContent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content ID"})
		return
	}

	var content models.Content

	if database.UseMemory {
		contentPtr, err := database.MemoryStore.GetContentByID(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Content not found"})
			return
		}
		content = *contentPtr
	} else {
		if err := database.DB.First(&content, uint(id)).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Content not found"})
			return
		}
	}

	// Check if user has permission to update this content
	_, exists := middleware.GetUserFromContext(c)
	if !exists || !middleware.HasPermission(c, content.UserID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var input struct {
		Title            string                `json:"title"`
		Summary          string                `json:"summary"`
		Type             models.ContentType    `json:"type"`
		InputType        models.ContentInputType `json:"input_type"`
		ContentData      string                `json:"content_data"`
		CoverImage       string                `json:"cover_image"`
		Tags             string                `json:"tags"`
		SourceURL        string                `json:"source_url"`
		Status           string                `json:"status"`
		Blocks           []models.ContentBlock `json:"blocks"`
		Category         string                `json:"category"`
		MetaTags         string                `json:"meta_tags"`
		GeneratedContent string                `json:"generated_content"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if input.Title != "" {
		content.Title = input.Title
	}
	if input.Summary != "" {
		content.Summary = input.Summary
	}
	if input.Type != "" {
		content.Type = input.Type
	}
	if input.InputType != "" {
		content.InputType = input.InputType
	}
	if input.ContentData != "" {
		content.ContentData = input.ContentData
	}
	if input.CoverImage != "" {
		content.CoverImage = input.CoverImage
	}
	if input.Tags != "" {
		content.Tags = input.Tags
	}
	if input.SourceURL != "" {
		content.SourceURL = input.SourceURL
	}
	if input.Status != "" {
		content.Status = input.Status
	}
	if input.Category != "" {
		content.Category = input.Category
	}
	if input.MetaTags != "" {
		content.MetaTags = input.MetaTags
	}
	if input.GeneratedContent != "" {
		content.GeneratedContent = input.GeneratedContent
	}

	if database.UseMemory {
		if err := database.MemoryStore.UpdateContent(&content); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update content"})
			return
		}
	} else {
		if err := database.DB.Save(&content).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update content"})
			return
		}

		// Update content blocks if provided
		if len(input.Blocks) > 0 {
			if err := database.DB.Where("content_id = ?", content.ID).Delete(&models.ContentBlock{}).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update content blocks"})
				return
			}

			for _, block := range input.Blocks {
				block.ContentID = content.ID
				if err := database.DB.Create(&block).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create content block"})
					return
				}
			}
		}

		// Fetch the updated content with associated blocks
		if err := database.DB.Preload("Blocks").First(&content, content.ID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated content"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Content updated successfully",
		"content": content,
	})
}

// DeleteContent deletes a content
func DeleteContent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content ID"})
		return
	}

	var content models.Content
	var userID uint

	if database.UseMemory {
		contentPtr, err := database.MemoryStore.GetContentByID(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Content not found"})
			return
		}
		content = *contentPtr
		userID = content.UserID
	} else {
		if err := database.DB.First(&content, uint(id)).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Content not found"})
			return
		}
		userID = content.UserID
	}

	// Check if user has permission to delete this content
	_, exists := middleware.GetUserFromContext(c)
	if !exists || !middleware.HasPermission(c, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if database.UseMemory {
		if err := database.MemoryStore.DeleteContent(uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete content"})
			return
		}
	} else {
		if err := database.DB.Delete(&content).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete content"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Content deleted successfully",
	})
}

// PublishContent publishes a content
func PublishContent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content ID"})
		return
	}

	var content models.Content

	if database.UseMemory {
		contentPtr, err := database.MemoryStore.GetContentByID(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Content not found"})
			return
		}
		content = *contentPtr
	} else {
		if err := database.DB.First(&content, uint(id)).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Content not found"})
			return
		}
	}

	// Check if user has permission to publish this content
	_, exists := middleware.GetUserFromContext(c)
	if !exists || !middleware.HasPermission(c, content.UserID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	content.Status = "published"

	if database.UseMemory {
		if err := database.MemoryStore.UpdateContent(&content); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish content"})
			return
		}
	} else {
		if err := database.DB.Save(&content).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish content"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Content published successfully",
		"content": content,
	})
}

// ArchiveContent archives a content
func ArchiveContent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content ID"})
		return
	}

	var content models.Content

	if database.UseMemory {
		contentPtr, err := database.MemoryStore.GetContentByID(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Content not found"})
			return
		}
		content = *contentPtr
	} else {
		if err := database.DB.First(&content, uint(id)).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Content not found"})
			return
		}
	}

	// Check if user has permission to archive this content
	_, exists := middleware.GetUserFromContext(c)
	if !exists || !middleware.HasPermission(c, content.UserID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	content.Status = "archived"

	if database.UseMemory {
		if err := database.MemoryStore.UpdateContent(&content); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive content"})
			return
		}
	} else {
		if err := database.DB.Save(&content).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive content"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Content archived successfully",
		"content": content,
	})
}

// GetContentStatistics returns statistics about contents
// GET /contents/statistics
func GetContentStatistics(c *gin.Context) {
	userID, exists := middleware.GetUserFromContext(c)
	role, _ := middleware.GetRoleFromContext(c)
	isAdmin := role == "admin"

	var contents []models.Content

	if database.UseMemory {
		memUserID := uint(0)
		if exists {
			memUserID = userID
		}
		contents = database.MemoryStore.GetContents(memUserID, isAdmin)
	} else {
		query := database.DB.Model(&models.Content{})
		if exists && !isAdmin {
			query = query.Where("user_id = ?", userID)
		}
		if err := query.Find(&contents).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch contents for statistics"})
			return
		}
	}

	total := len(contents)
	draft := 0
	generated := 0
	published := 0
	archived := 0

	for _, c := range contents {
		switch c.Status {
		case "draft":
			draft++
		case "published":
			published++
		case "archived":
			archived++
		}
		if c.GeneratedContent != "" || c.Status != "draft" {
			generated++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     total,
		"draft":     draft,
		"generated": generated,
		"published": published,
		"archived":  archived,
	})
}

// GenerateContentWithAI generates content using AI service
// POST /ai/generate
// Request body: { prompt, type, input_type, url, file_name }
// Response: { content: Content object (saved to DB), generated_title, summary }
func GenerateContentWithAI(c *gin.Context) {
	userID, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input struct {
		Prompt    string                 `json:"prompt"`
		Type      models.ContentType     `json:"type" binding:"required"`
		InputType models.ContentInputType `json:"input_type" binding:"required"`
		URL       string                 `json:"url"`
		FileName  string                 `json:"file_name"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var generatedBody string
	var err error
	var promptForAI string

	switch input.InputType {
	case models.PromptInput:
		promptForAI = input.Prompt
		generatedBody, err = aiService.GenerateTextContent(input.Prompt, input.Type)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate content from prompt"})
			return
		}
	case models.LinkInput:
		promptForAI = input.URL
		generatedBody, err = aiService.ExtractContentFromURL(input.URL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract content from URL"})
			return
		}
	case models.FileInput:
		promptForAI = input.FileName
		generatedBody, err = aiService.ExtractContentFromFile([]byte{}, input.FileName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract content from file"})
			return
		}
	case models.PasteInput:
		// User pasted content directly in prompt field
		promptForAI = input.Prompt
		generatedBody = input.Prompt
	default:
		promptForAI = input.Prompt
		generatedBody, err = aiService.GenerateTextContent(input.Prompt, input.Type)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate content"})
			return
		}
	}

	// Generate structured fields
	title, _ := aiService.GenerateTitleFromPrompt(promptForAI)
	summary, _ := aiService.GenerateSummary(promptForAI, generatedBody)
	metaTags, _ := aiService.GenerateMetaTags(promptForAI)
	category, _ := aiService.DetectCategory(promptForAI)

	content := models.Content{
		UserID:           userID,
		Title:            title,
		Summary:          summary,
		Type:             input.Type,
		InputType:        input.InputType,
		ContentData:      generatedBody,
		GeneratedContent: generatedBody,
		Status:           "draft",
		Category:         category,
		MetaTags:         metaTags,
		SourceURL:        input.URL,
	}

	// Save to database
	if database.UseMemory {
		if err := database.MemoryStore.CreateContent(&content); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save generated content"})
			return
		}
	} else {
		if err := database.DB.Create(&content).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save generated content"})
			return
		}
		if err := database.DB.First(&content, content.ID).Error; err == nil {
			// Reloaded
		}
	}

	// 生成结构化内容块 (blocks) 并保存
	structured, err := aiService.GenerateStructuredContent(promptForAI, input.Type)
	if err == nil && structured != nil {
		content.Summary = summary
		if content.Summary == "" {
			content.Summary = structured.Summary
		}
		var blocks []models.ContentBlock
		for i := range structured.Blocks {
			block := models.ContentBlock{
				ContentID: content.ID,
				BlockType: structured.Blocks[i].BlockType,
				Content:   structured.Blocks[i].Content,
				MediaURL:  structured.Blocks[i].MediaURL,
				Order:     structured.Blocks[i].Order,
			}
			if database.UseMemory {
				_ = database.MemoryStore.CreateContentBlock(&block)
			} else {
				_ = database.DB.Create(&block).Error
			}
			blocks = append(blocks, block)
		}
		content.Blocks = blocks
	}

	c.JSON(http.StatusOK, gin.H{
		"content":         content,
		"generated_title": title,
		"summary":         summary,
		"meta_tags":       metaTags,
		"category":        category,
	})
}

// GenerateContentBlocks generates structured content blocks for an existing content
// POST /ai/generate-blocks/:id
func GenerateContentBlocks(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content ID"})
		return
	}

	var content models.Content

	if database.UseMemory {
		contentPtr, err := database.MemoryStore.GetContentByID(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Content not found"})
			return
		}
		content = *contentPtr
	} else {
		if err := database.DB.First(&content, uint(id)).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Content not found"})
			return
		}
	}

	// Check permission
	_, exists := middleware.GetUserFromContext(c)
	if !exists || !middleware.HasPermission(c, content.UserID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Build prompt for block generation
	basePrompt := content.Title
	if content.ContentData != "" {
		basePrompt = content.Title + " " + content.ContentData
	} else if content.GeneratedContent != "" {
		basePrompt = content.Title + " " + content.GeneratedContent
	}

	// Use AI service to generate structured content
	structured, err := aiService.GenerateStructuredContent(basePrompt, content.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate structured content"})
		return
	}

	blocks := structured.Blocks

	// Attach content ID to all blocks
	for i := range blocks {
		blocks[i].ContentID = content.ID
	}

	// Persist blocks
	if database.UseMemory {
		for i := range blocks {
			_ = database.MemoryStore.CreateContentBlock(&blocks[i])
		}
	} else {
		// Remove old blocks
		_ = database.DB.Where("content_id = ?", content.ID).Delete(&models.ContentBlock{}).Error
		// Create new blocks
		for i := range blocks {
			_ = database.DB.Create(&blocks[i]).Error
		}
	}

	// Update content with structured information
	if content.Category == "" {
		category, _ := aiService.DetectCategory(basePrompt)
		content.Category = category
	}
	if content.MetaTags == "" {
		metaTags, _ := aiService.GenerateMetaTags(basePrompt)
		content.MetaTags = metaTags
	}
	if content.Summary == "" {
		summary, _ := aiService.GenerateSummary(basePrompt, content.ContentData)
		content.Summary = summary
	}

	if database.UseMemory {
		_ = database.MemoryStore.UpdateContent(&content)
	} else {
		_ = database.DB.Save(&content).Error
	}

	c.JSON(http.StatusOK, gin.H{
		"blocks":  blocks,
		"content": content,
		"_debug":  json.RawMessage(`{"ok":true}`),
	})
}
