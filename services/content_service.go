package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ai-content-platform/database"
	"ai-content-platform/models"

	"github.com/jinzhu/gorm"
)

// ContentService provides methods for managing content
type ContentService struct{}

// CreateContent creates a new content with associated blocks
func (cs *ContentService) CreateContent(content *models.Content) error {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(content).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Create content blocks if provided
	for i := range content.Blocks {
		content.Blocks[i].ContentID = content.ID
		if err := tx.Create(&content.Blocks[i]).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// GetContent retrieves a content by ID with its associated blocks and distributions
func (cs *ContentService) GetContent(id uint) (*models.Content, error) {
	var content models.Content
	err := database.DB.Preload("User").Preload("Blocks").Preload("Distributors").First(&content, id).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.New("content not found")
		}
		return nil, err
	}
	return &content, nil
}

// GetContents retrieves contents with optional filters
func (cs *ContentService) GetContents(userID *uint, contentType *models.ContentType, status *string, title *string, limit, offset int) ([]models.Content, int, error) {
	query := database.DB.Preload("User").Preload("Blocks").Preload("Distributors").Model(&models.Content{})

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if contentType != nil {
		query = query.Where("type = ?", *contentType)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if title != nil {
		query = query.Where("title LIKE ?", fmt.Sprintf("%%%s%%", *title))
	}

	var total int
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var contents []models.Content
	if err := query.Offset(offset).Limit(limit).Find(&contents).Error; err != nil {
		return nil, 0, err
	}

	return contents, total, nil
}

// UpdateContent updates an existing content
func (cs *ContentService) UpdateContent(content *models.Content) error {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update the main content
	if err := tx.Save(content).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete existing blocks
	if err := tx.Where("content_id = ?", content.ID).Delete(&models.ContentBlock{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Create new blocks
	for i := range content.Blocks {
		content.Blocks[i].ContentID = content.ID
		if err := tx.Create(&content.Blocks[i]).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// DeleteContent deletes a content by ID
func (cs *ContentService) DeleteContent(id uint) error {
	content := models.Content{ID: id}
	return database.DB.Delete(&content).Error
}

// PublishContent publishes a content
func (cs *ContentService) PublishContent(id uint) error {
	content, err := cs.GetContent(id)
	if err != nil {
		return err
	}

	content.Status = "published"
	return database.DB.Save(content).Error
}

// ArchiveContent archives a content
func (cs *ContentService) ArchiveContent(id uint) error {
	content, err := cs.GetContent(id)
	if err != nil {
		return err
	}

	content.Status = "archived"
	return database.DB.Save(content).Error
}

// DistributeContentToPlatform distributes content to a specific platform
func (cs *ContentService) DistributeContentToPlatform(contentID uint, platform string) error {
	content, err := cs.GetContent(contentID)
	if err != nil {
		return err
	}
	_ = content // content is retrieved for validation, not directly modified here

	// Create distribution record
	distribution := models.ContentDistribution{
		ContentID: contentID,
		Platform:  platform,
		Status:    "pending",
	}

	if err := database.DB.Create(&distribution).Error; err != nil {
		return err
	}

	// Simulate async distribution process
	go cs.processDistribution(distribution.ID)

	return nil
}

// processDistribution handles the actual distribution to platforms
func (cs *ContentService) processDistribution(distributionID uint) {
	time.Sleep(2 * time.Second) // Simulate processing time

	var distribution models.ContentDistribution
	if err := database.DB.First(&distribution, distributionID).Error; err != nil {
		return
	}

	// In a real implementation, this would call the appropriate API for the platform
	// For now, we'll simulate success or failure randomly
	distribution.Status = "success"
	distribution.PublicationDate = &time.Time{}
	distribution.URL = fmt.Sprintf("https://example.com/%s/%d", distribution.Platform, distribution.ContentID)

	// Update distribution status
	database.DB.Save(&distribution)
}

// GenerateContentFromPrompt generates content based on a prompt
func (cs *ContentService) GenerateContentFromPrompt(prompt string, contentType models.ContentType) (*models.Content, error) {
	// This is a simplified simulation of AI content generation
	// In a real implementation, this would call an AI API

	var contentData string
	var title string

	switch contentType {
	case models.Article:
		title = "AI Generated Article: " + prompt[:min(len(prompt), 30)]
		contentData = fmt.Sprintf(`{
			"title": "%s",
			"sections": [
				{
					"type": "paragraph",
					"content": "This is an AI-generated article based on the prompt: '%s'. The article explores key concepts and provides insights.",
					"order": 1
				},
				{
					"type": "paragraph", 
					"content": "Additional content generated by AI to fulfill the user's request. This section expands on the main ideas.",
					"order": 2
				}
			]
		}`, title, prompt)
	case models.Poster:
		title = "AI Generated Poster: " + prompt[:min(len(prompt), 30)]
		contentData = fmt.Sprintf(`{
			"title": "%s",
			"design_elements": [
				{
					"type": "text",
					"content": "Main heading based on: %s",
					"position": {"x": 50, "y": 100}
				},
				{
					"type": "text",
					"content": "Subheading with key message",
					"position": {"x": 50, "y": 150}
				}
			]
		}`, title, prompt)
	default:
		title = "AI Generated Content: " + prompt[:min(len(prompt), 30)]
		contentData = fmt.Sprintf(`{"prompt": "%s", "generated_content": "AI-generated content based on the provided prompt"}`, prompt)
	}

	content := &models.Content{
		Title:       title,
		Type:        contentType,
		InputType:   models.PromptInput,
		ContentData: contentData,
		Status:      "draft",
	}

	// Create content blocks based on generated content
	var blocks []models.ContentBlock
	if contentType == models.Article {
		// Parse the contentData to extract blocks
		var articleData map[string]interface{}
		if err := json.Unmarshal([]byte(contentData), &articleData); err == nil {
			if sections, ok := articleData["sections"].([]interface{}); ok {
				for i, section := range sections {
					if secMap, ok := section.(map[string]interface{}); ok {
						blockType := models.ParagraphBlock
						if t, ok := secMap["type"].(string); ok {
							switch t {
							case "title":
								blockType = models.TitleBlock
							case "paragraph":
								blockType = models.ParagraphBlock
							case "image":
								blockType = models.ImageBlock
							case "list":
								blockType = models.ListBlock
							case "table":
								blockType = models.TableBlock
							}
						}

						contentStr := ""
						if c, ok := secMap["content"].(string); ok {
							contentStr = c
						}

						order := i + 1
						if o, ok := secMap["order"].(float64); ok {
							order = int(o)
						}

						blocks = append(blocks, models.ContentBlock{
							BlockType: blockType,
							Content:   contentStr,
							Order:     order,
						})
					}
				}
			}
		}
	}

	content.Blocks = blocks

	return content, nil
}

// min is a helper function to find the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
