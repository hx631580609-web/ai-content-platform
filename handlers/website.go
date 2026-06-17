package handlers

import (
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"ai-content-platform/database"
	"ai-content-platform/models"

	"github.com/gin-gonic/gin"
)

// GetNavItems retrieves all navigation items
func GetNavItems(c *gin.Context) {
	if database.UseMemory {
		items := database.MemoryStore.GetEnabledNavItems()
		sort.Slice(items, func(i, j int) bool {
			return items[i].Position < items[j].Position
		})
		c.JSON(http.StatusOK, gin.H{
			"nav_items": items,
		})
		return
	}

	var items []models.NavItem
	if err := database.DB.Where("enabled = ? AND parent_id = ?", true, 0).Order("position").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch navigation items"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"nav_items": items,
	})
}

// GetAllNavItems retrieves all navigation items including disabled ones (for admin)
func GetAllNavItems(c *gin.Context) {
	if database.UseMemory {
		items := database.MemoryStore.GetAllNavItems()
		sort.Slice(items, func(i, j int) bool {
			return items[i].Position < items[j].Position
		})
		c.JSON(http.StatusOK, gin.H{
			"nav_items": items,
		})
		return
	}

	var items []models.NavItem
	if err := database.DB.Order("position").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch navigation items"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"nav_items": items,
	})
}

// GetNavItem retrieves a specific navigation item by slug
func GetNavItem(c *gin.Context) {
	slug := c.Param("slug")

	if database.UseMemory {
		item, err := database.MemoryStore.GetNavItemBySlug(slug)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Navigation item not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"nav_item": item,
		})
		return
	}

	var item models.NavItem
	if err := database.DB.Where("slug = ?", slug).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Navigation item not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"nav_item": item,
	})
}

// CreateNavItem creates a new navigation item
func CreateNavItem(c *gin.Context) {
	var input struct {
		Name        string `json:"name" binding:"required"`
		NameEn      string `json:"name_en"`
		Slug        string `json:"slug" binding:"required"`
		NavType     string `json:"nav_type" binding:"required"`
		LinkType    string `json:"link_type"`
		URL         string `json:"url"`
		ServiceType string `json:"service_type"`
		Position    int    `json:"position"`
		Enabled     bool   `json:"enabled"`
		ParentID    uint   `json:"parent_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if database.UseMemory {
		if database.MemoryStore.NavSlugExists(input.Slug) {
			c.JSON(http.StatusConflict, gin.H{"error": "Navigation item with this slug already exists"})
			return
		}

		navItem := models.NavItem{
			Name:        input.Name,
			NameEn:      input.NameEn,
			Slug:        input.Slug,
			NavType:     input.NavType,
			LinkType:    input.LinkType,
			URL:         input.URL,
			ServiceType: input.ServiceType,
			Position:    input.Position,
			Enabled:     input.Enabled,
			ParentID:    input.ParentID,
		}

		if err := database.MemoryStore.CreateNavItem(&navItem); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create navigation item"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":  "Navigation item created successfully",
			"nav_item": navItem,
		})
		return
	}

	var existing models.NavItem
	if database.DB.Where("slug = ?", input.Slug).First(&existing).Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Navigation item with this slug already exists"})
		return
	}

	navItem := models.NavItem{
		Name:        input.Name,
		NameEn:      input.NameEn,
		Slug:        input.Slug,
		NavType:     input.NavType,
		LinkType:    input.LinkType,
		URL:         input.URL,
		ServiceType: input.ServiceType,
		Position:    input.Position,
		Enabled:     input.Enabled,
		ParentID:    input.ParentID,
	}

	if err := database.DB.Create(&navItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create navigation item"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Navigation item created successfully",
		"nav_item": navItem,
	})
}

// UpdateNavItem updates an existing navigation item
func UpdateNavItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid navigation item ID"})
		return
	}

	var input struct {
		Name        string `json:"name"`
		NameEn      string `json:"name_en"`
		Slug        string `json:"slug"`
		NavType     string `json:"nav_type"`
		LinkType    string `json:"link_type"`
		URL         string `json:"url"`
		ServiceType string `json:"service_type"`
		Position    int    `json:"position"`
		Enabled     bool   `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if database.UseMemory {
		item, err := database.MemoryStore.GetNavItemByID(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Navigation item not found"})
			return
		}

		updated := *item
		if input.Name != "" {
			updated.Name = input.Name
		}
		if input.NameEn != "" {
			updated.NameEn = input.NameEn
		}
		if input.Slug != "" {
			if input.Slug != item.Slug && database.MemoryStore.NavSlugExists(input.Slug) {
				c.JSON(http.StatusConflict, gin.H{"error": "Navigation item with this slug already exists"})
				return
			}
			updated.Slug = input.Slug
		}
		if input.NavType != "" {
			updated.NavType = input.NavType
		}
		if input.LinkType != "" {
			updated.LinkType = input.LinkType
		}
		if input.URL != "" {
			updated.URL = input.URL
		}
		if input.ServiceType != "" {
			updated.ServiceType = input.ServiceType
		}
		if input.Position != 0 || input.Name != "" {
			updated.Position = input.Position
		}
		updated.Enabled = input.Enabled

		if err := database.MemoryStore.UpdateNavItem(&updated); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update navigation item"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "Navigation item updated successfully",
			"nav_item": updated,
		})
		return
	}

	var navItem models.NavItem
	if err := database.DB.First(&navItem, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Navigation item not found"})
		return
	}

	if input.Name != "" {
		navItem.Name = input.Name
	}
	if input.NameEn != "" {
		navItem.NameEn = input.NameEn
	}
	if input.Slug != "" {
		var existing models.NavItem
		if database.DB.Where("slug = ? AND id != ?", input.Slug, navItem.ID).First(&existing).Error == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Navigation item with this slug already exists"})
			return
		}
		navItem.Slug = input.Slug
	}
	if input.NavType != "" {
		navItem.NavType = input.NavType
	}
	if input.LinkType != "" {
		navItem.LinkType = input.LinkType
	}
	if input.URL != "" {
		navItem.URL = input.URL
	}
	if input.ServiceType != "" {
		navItem.ServiceType = input.ServiceType
	}
	if input.Position != 0 || input.Name != "" {
		navItem.Position = input.Position
	}
	navItem.Enabled = input.Enabled

	if err := database.DB.Save(&navItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update navigation item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Navigation item updated successfully",
		"nav_item": navItem,
	})
}

// DeleteNavItem deletes a navigation item
func DeleteNavItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid navigation item ID"})
		return
	}

	if database.UseMemory {
		if err := database.MemoryStore.DeleteNavItem(uint(id)); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Navigation item not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Navigation item deleted successfully",
		})
		return
	}

	var navItem models.NavItem
	if err := database.DB.First(&navItem, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Navigation item not found"})
		return
	}

	if err := database.DB.Delete(&navItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete navigation item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Navigation item deleted successfully",
	})
}

// GetWebsiteModules retrieves all website modules
func GetWebsiteModules(c *gin.Context) {
	var modules []models.WebsiteModule
	if err := database.DB.Find(&modules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch website modules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"modules": modules,
	})
}

// GetWebsiteModule retrieves a specific website module by name
func GetWebsiteModule(c *gin.Context) {
	name := c.Param("name")

	var module models.WebsiteModule
	if err := database.DB.Where("name = ?", name).First(&module).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Website module not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"module": module,
	})
}

// UpdateWebsiteModule updates a website module
func UpdateWebsiteModule(c *gin.Context) {
	name := c.Param("name")

	var module models.WebsiteModule
	if err := database.DB.Where("name = ?", name).First(&module).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Website module not found"})
		return
	}

	var input struct {
		Enabled  bool   `json:"enabled"`
		Position int    `json:"position"`
		Config   string `json:"config"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	module.Enabled = input.Enabled
	module.Position = input.Position
	if input.Config != "" {
		module.Config = input.Config
	}

	if err := database.DB.Save(&module).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update website module"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Website module updated successfully",
		"module":  module,
	})
}

// GetBlogPosts retrieves all blog posts with pagination and filtering
func GetBlogPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	category := c.Query("category")
	status := c.Query("status")
	search := c.Query("search")

	if database.UseMemory {
		allBlogs := database.MemoryStore.GetAllBlogPosts()
		var filtered []models.BlogPost
		for _, blog := range allBlogs {
			if category != "" && blog.Category != category {
				continue
			}
			if status != "" && blog.Status != status {
				continue
			}
			if search != "" {
				lowerSearch := strings.ToLower(search)
				if !strings.Contains(strings.ToLower(blog.Title), lowerSearch) &&
					!strings.Contains(strings.ToLower(blog.Content), lowerSearch) {
					continue
				}
			}
			if blog.AuthorID > 0 {
				if author, err := database.MemoryStore.GetUserByID(blog.AuthorID); err == nil {
					authorCopy := *author
					authorCopy.Password = ""
					blog.Author = authorCopy
				}
			}
			filtered = append(filtered, blog)
		}

		total := int64(len(filtered))
		if offset >= len(filtered) {
			filtered = []models.BlogPost{}
		} else {
			end := offset + limit
			if end > len(filtered) {
				end = len(filtered)
			}
			filtered = filtered[offset:end]
		}

		c.JSON(http.StatusOK, gin.H{
			"blog_posts": filtered,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		})
		return
	}

	query := database.DB.Preload("Author", "id, username, email").Model(&models.BlogPost{})

	// Apply filters
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var blogPosts []models.BlogPost
	var total int64

	if err := query.Offset(offset).Limit(limit).Find(&blogPosts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch blog posts"})
		return
	}

	query.Model(&models.BlogPost{}).Count(&total)

	c.JSON(http.StatusOK, gin.H{
		"blog_posts": blogPosts,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetBlogPost retrieves a specific blog post by slug
func GetBlogPost(c *gin.Context) {
	slug := c.Param("slug")

	if database.UseMemory {
		blog, err := database.MemoryStore.GetBlogPostBySlug(slug)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Blog post not found"})
			return
		}

		// Increment view count
		blog.ViewCount++
		if err := database.MemoryStore.UpdateBlogPost(blog); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update blog post"})
			return
		}

		if blog.AuthorID > 0 {
			if author, err := database.MemoryStore.GetUserByID(blog.AuthorID); err == nil {
				authorCopy := *author
				authorCopy.Password = ""
				blog.Author = authorCopy
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"blog_post": blog,
		})
		return
	}

	var blogPost models.BlogPost
	if err := database.DB.Preload("Author", "id, username, email").First(&blogPost, "slug = ?", slug).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Blog post not found"})
		return
	}

	// Increment view count
	blogPost.ViewCount++
	database.DB.Save(&blogPost)

	c.JSON(http.StatusOK, gin.H{
		"blog_post": blogPost,
	})
}

// CreateBlogPost creates a new blog post
func CreateBlogPost(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input struct {
		Title         string  `json:"title" binding:"required"`
		Slug          string  `json:"slug" binding:"required"`
		Content       string  `json:"content" binding:"required"`
		Summary       string  `json:"summary"`
		Category      string  `json:"category"`
		CoverImage    string  `json:"cover_image"`
		Status        string  `json:"status"`
		PublishedAt   *string `json:"published_at"` // ISO string format
		IsAiGenerated bool    `json:"is_ai_generated"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if database.UseMemory {
		if database.MemoryStore.BlogSlugExists(input.Slug) {
			c.JSON(http.StatusConflict, gin.H{"error": "Blog post with this slug already exists"})
			return
		}

		blogPost := models.BlogPost{
			Title:         input.Title,
			Slug:          input.Slug,
			Content:       input.Content,
			Summary:       input.Summary,
			Category:      input.Category,
			CoverImage:    input.CoverImage,
			AuthorID:      userID.(uint),
			Status:        input.Status,
			IsAiGenerated: input.IsAiGenerated,
		}

		if blogPost.Status == "" {
			blogPost.Status = "draft"
		}

		if err := database.MemoryStore.CreateBlogPost(&blogPost); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create blog post"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":   "Blog post created successfully",
			"blog_post": blogPost,
		})
		return
	}

	// Check if slug already exists
	var existingPost models.BlogPost
	if database.DB.Where("slug = ?", input.Slug).First(&existingPost).Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Blog post with this slug already exists"})
		return
	}

	blogPost := models.BlogPost{
		Title:         input.Title,
		Slug:          input.Slug,
		Content:       input.Content,
		Summary:       input.Summary,
		Category:      input.Category,
		CoverImage:    input.CoverImage,
		AuthorID:      userID.(uint),
		Status:        input.Status,
		IsAiGenerated: input.IsAiGenerated,
	}

	// Set default status to draft if not provided
	if blogPost.Status == "" {
		blogPost.Status = "draft"
	}

	if err := database.DB.Create(&blogPost).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create blog post"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Blog post created successfully",
		"blog_post": blogPost,
	})
}

// UpdateBlogPost updates an existing blog post
func UpdateBlogPost(c *gin.Context) {
	slug := c.Param("slug")

	// Check if user has permission to update this blog post
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	role, _ := c.Get("role")
	isAdmin := false
	if role != nil {
		if roleStr, ok := role.(string); ok && roleStr == "admin" {
			isAdmin = true
		}
	}

	var input struct {
		Title       string  `json:"title"`
		Slug        string  `json:"slug"`
		Content     string  `json:"content"`
		Summary     string  `json:"summary"`
		Category    string  `json:"category"`
		CoverImage  string  `json:"cover_image"`
		Status      string  `json:"status"`
		PublishedAt *string `json:"published_at"` // ISO string format
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if database.UseMemory {
		blog, err := database.MemoryStore.GetBlogPostBySlug(slug)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Blog post not found"})
			return
		}

		if userID.(uint) != blog.AuthorID && !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		updated := *blog
		if input.Title != "" {
			updated.Title = input.Title
		}
		if input.Slug != "" {
			if input.Slug != slug && database.MemoryStore.BlogSlugExists(input.Slug) {
				c.JSON(http.StatusConflict, gin.H{"error": "Blog post with this slug already exists"})
				return
			}
			updated.Slug = input.Slug
		}
		if input.Content != "" {
			updated.Content = input.Content
		}
		if input.Summary != "" {
			updated.Summary = input.Summary
		}
		if input.Category != "" {
			updated.Category = input.Category
		}
		if input.CoverImage != "" {
			updated.CoverImage = input.CoverImage
		}
		if input.Status != "" {
			updated.Status = input.Status
		}

		if err := database.MemoryStore.UpdateBlogPost(&updated); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update blog post"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":   "Blog post updated successfully",
			"blog_post": updated,
		})
		return
	}

	var blogPost models.BlogPost
	if err := database.DB.First(&blogPost, "slug = ?", slug).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Blog post not found"})
		return
	}

	if userID.(uint) != blogPost.AuthorID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Update fields if provided
	if input.Title != "" {
		blogPost.Title = input.Title
	}
	if input.Slug != "" {
		// Check if new slug already exists
		var existingPost models.BlogPost
		if database.DB.Where("slug = ? AND id != ?", input.Slug, blogPost.ID).First(&existingPost).Error == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Blog post with this slug already exists"})
			return
		}
		blogPost.Slug = input.Slug
	}
	if input.Content != "" {
		blogPost.Content = input.Content
	}
	if input.Summary != "" {
		blogPost.Summary = input.Summary
	}
	if input.Category != "" {
		blogPost.Category = input.Category
	}
	if input.CoverImage != "" {
		blogPost.CoverImage = input.CoverImage
	}
	if input.Status != "" {
		blogPost.Status = input.Status
	}

	if err := database.DB.Save(&blogPost).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update blog post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Blog post updated successfully",
		"blog_post": blogPost,
	})
}

// DeleteBlogPost deletes a blog post
func DeleteBlogPost(c *gin.Context) {
	slug := c.Param("slug")

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	role, _ := c.Get("role")
	isAdmin := false
	if role != nil {
		if roleStr, ok := role.(string); ok && roleStr == "admin" {
			isAdmin = true
		}
	}

	if database.UseMemory {
		blog, err := database.MemoryStore.GetBlogPostBySlug(slug)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Blog post not found"})
			return
		}

		if userID.(uint) != blog.AuthorID && !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		if err := database.MemoryStore.DeleteBlogPost(blog.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete blog post"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Blog post deleted successfully",
		})
		return
	}

	var blogPost models.BlogPost
	if err := database.DB.First(&blogPost, "slug = ?", slug).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Blog post not found"})
		return
	}

	// Check if user has permission to delete this blog post
	if userID.(uint) != blogPost.AuthorID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := database.DB.Delete(&blogPost).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete blog post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Blog post deleted successfully",
	})
}

// GetServicePages retrieves all service pages
func GetServicePages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	serviceType := c.Query("service_type")
	status := c.Query("status")

	if database.UseMemory {
		allPages := database.MemoryStore.GetAllServicePages()
		var filtered []models.ServicePage
		for _, sp := range allPages {
			if serviceType != "" && sp.ServiceType != serviceType {
				continue
			}
			if status != "" && sp.Status != status {
				continue
			}
			filtered = append(filtered, sp)
		}

		total := int64(len(filtered))
		if offset >= len(filtered) {
			filtered = []models.ServicePage{}
		} else {
			end := offset + limit
			if end > len(filtered) {
				end = len(filtered)
			}
			filtered = filtered[offset:end]
		}

		c.JSON(http.StatusOK, gin.H{
			"service_pages": filtered,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		})
		return
	}

	query := database.DB.Model(&models.ServicePage{})

	// Apply filters
	if serviceType != "" {
		query = query.Where("service_type = ?", serviceType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var servicePages []models.ServicePage
	var total int64

	if err := query.Offset(offset).Limit(limit).Find(&servicePages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch service pages"})
		return
	}

	query.Model(&models.ServicePage{}).Count(&total)

	c.JSON(http.StatusOK, gin.H{
		"service_pages": servicePages,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetServicePage retrieves a specific service page by slug
func GetServicePage(c *gin.Context) {
	slug := c.Param("slug")

	if database.UseMemory {
		page, err := database.MemoryStore.GetServicePageBySlug(slug)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service page not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"service_page": page,
		})
		return
	}

	var servicePage models.ServicePage
	if err := database.DB.First(&servicePage, "slug = ?", slug).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service page not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service_page": servicePage,
	})
}

// CreateServicePage creates a new service page
func CreateServicePage(c *gin.Context) {
	var input struct {
		Title       string `json:"title" binding:"required"`
		Slug        string `json:"slug" binding:"required"`
		Description string `json:"description"`
		Content     string `json:"content" binding:"required"`
		ServiceType string `json:"service_type" binding:"required"`
		CoverImage  string `json:"cover_image"`
		Status      string `json:"status"`
		ContactInfo string `json:"contact_info"`
		Faq         string `json:"faq"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if database.UseMemory {
		if database.MemoryStore.ServicePageSlugExists(input.Slug) {
			c.JSON(http.StatusConflict, gin.H{"error": "Service page with this slug already exists"})
			return
		}

		servicePage := models.ServicePage{
			Title:       input.Title,
			Slug:        input.Slug,
			Description: input.Description,
			Content:     input.Content,
			ServiceType: input.ServiceType,
			CoverImage:  input.CoverImage,
			Status:      input.Status,
			ContactInfo: input.ContactInfo,
			Faq:         input.Faq,
		}

		if servicePage.Status == "" {
			servicePage.Status = "active"
		}

		if err := database.MemoryStore.CreateServicePage(&servicePage); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create service page"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":      "Service page created successfully",
			"service_page": servicePage,
		})
		return
	}

	// Check if slug already exists
	var existingPage models.ServicePage
	if database.DB.Where("slug = ?", input.Slug).First(&existingPage).Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Service page with this slug already exists"})
		return
	}

	servicePage := models.ServicePage{
		Title:       input.Title,
		Slug:        input.Slug,
		Description: input.Description,
		Content:     input.Content,
		ServiceType: input.ServiceType,
		CoverImage:  input.CoverImage,
		Status:      input.Status,
		ContactInfo: input.ContactInfo,
		Faq:         input.Faq,
	}

	// Set default status to active if not provided
	if servicePage.Status == "" {
		servicePage.Status = "active"
	}

	if err := database.DB.Create(&servicePage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create service page"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Service page created successfully",
		"service_page": servicePage,
	})
}

// UpdateServicePage updates an existing service page
func UpdateServicePage(c *gin.Context) {
	slug := c.Param("slug")

	var input struct {
		Title       string `json:"title"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		Content     string `json:"content"`
		ServiceType string `json:"service_type"`
		CoverImage  string `json:"cover_image"`
		Status      string `json:"status"`
		ContactInfo string `json:"contact_info"`
		Faq         string `json:"faq"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if database.UseMemory {
		existing, err := database.MemoryStore.GetServicePageBySlug(slug)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service page not found"})
			return
		}

		updated := *existing
		if input.Title != "" {
			updated.Title = input.Title
		}
		if input.Slug != "" {
			if input.Slug != slug && database.MemoryStore.ServicePageSlugExists(input.Slug) {
				c.JSON(http.StatusConflict, gin.H{"error": "Service page with this slug already exists"})
				return
			}
			updated.Slug = input.Slug
		}
		if input.Description != "" {
			updated.Description = input.Description
		}
		if input.Content != "" {
			updated.Content = input.Content
		}
		if input.ServiceType != "" {
			updated.ServiceType = input.ServiceType
		}
		if input.CoverImage != "" {
			updated.CoverImage = input.CoverImage
		}
		if input.Status != "" {
			updated.Status = input.Status
		}
		if input.ContactInfo != "" {
			updated.ContactInfo = input.ContactInfo
		}
		if input.Faq != "" {
			updated.Faq = input.Faq
		}

		if err := database.MemoryStore.UpdateServicePage(&updated); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update service page"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "Service page updated successfully",
			"service_page": updated,
		})
		return
	}

	var servicePage models.ServicePage
	if err := database.DB.First(&servicePage, "slug = ?", slug).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service page not found"})
		return
	}

	// Update fields if provided
	if input.Title != "" {
		servicePage.Title = input.Title
	}
	if input.Slug != "" {
		// Check if new slug already exists
		var existingPage models.ServicePage
		if database.DB.Where("slug = ? AND id != ?", input.Slug, servicePage.ID).First(&existingPage).Error == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Service page with this slug already exists"})
			return
		}
		servicePage.Slug = input.Slug
	}
	if input.Description != "" {
		servicePage.Description = input.Description
	}
	if input.Content != "" {
		servicePage.Content = input.Content
	}
	if input.ServiceType != "" {
		servicePage.ServiceType = input.ServiceType
	}
	if input.CoverImage != "" {
		servicePage.CoverImage = input.CoverImage
	}
	if input.Status != "" {
		servicePage.Status = input.Status
	}
	if input.ContactInfo != "" {
		servicePage.ContactInfo = input.ContactInfo
	}
	if input.Faq != "" {
		servicePage.Faq = input.Faq
	}

	if err := database.DB.Save(&servicePage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update service page"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Service page updated successfully",
		"service_page": servicePage,
	})
}

// DeleteServicePage deletes a service page
func DeleteServicePage(c *gin.Context) {
	slug := c.Param("slug")

	if database.UseMemory {
		if !database.MemoryStore.ServicePageSlugExists(slug) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service page not found"})
			return
		}

		if err := database.MemoryStore.DeleteServicePage(slug); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete service page"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Service page deleted successfully",
		})
		return
	}

	var servicePage models.ServicePage
	if err := database.DB.First(&servicePage, "slug = ?", slug).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service page not found"})
		return
	}

	if err := database.DB.Delete(&servicePage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete service page"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Service page deleted successfully",
	})
}

// GetAboutUs retrieves the about us page content
func GetAboutUs(c *gin.Context) {
	var aboutUs models.AboutUs
	if err := database.DB.First(&aboutUs).Error; err != nil {
		// If no record exists, return default empty content
		c.JSON(http.StatusOK, gin.H{
			"about_us": models.AboutUs{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"about_us": aboutUs,
	})
}

// UpdateAboutUs updates the about us page content
func UpdateAboutUs(c *gin.Context) {
	var aboutUs models.AboutUs

	// Try to find existing record, if not found, create a new one
	if err := database.DB.First(&aboutUs).Error; err != nil {
		// Create new record if not found
		aboutUs = models.AboutUs{}
	}

	var input struct {
		CompanyIntro   string `json:"company_intro"`
		Qualifications string `json:"qualifications"`
		TeamPhotos     string `json:"team_photos"`
		ContactInfo    string `json:"contact_info"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if input.CompanyIntro != "" {
		aboutUs.CompanyIntro = input.CompanyIntro
	}
	if input.Qualifications != "" {
		aboutUs.Qualifications = input.Qualifications
	}
	if input.TeamPhotos != "" {
		aboutUs.TeamPhotos = input.TeamPhotos
	}
	if input.ContactInfo != "" {
		aboutUs.ContactInfo = input.ContactInfo
	}

	if err := database.DB.Save(&aboutUs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update about us page"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "About us page updated successfully",
		"about_us": aboutUs,
	})
}

// GetFooterContact retrieves the footer contact information
func GetFooterContact(c *gin.Context) {
	var footerContact models.FooterContact
	if err := database.DB.First(&footerContact).Error; err != nil {
		// If no record exists, return default empty content
		c.JSON(http.StatusOK, gin.H{
			"footer_contact": models.FooterContact{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"footer_contact": footerContact,
	})
}

// UpdateFooterContact updates the footer contact information
func UpdateFooterContact(c *gin.Context) {
	var footerContact models.FooterContact

	// Try to find existing record, if not found, create a new one
	if err := database.DB.First(&footerContact).Error; err != nil {
		// Create new record if not found
		footerContact = models.FooterContact{}
	}

	var input struct {
		WeChatQR string `json:"wechat_qr"`
		WeChatID string `json:"wechat_id"`
		Phone    string `json:"phone"`
		Email    string `json:"email"`
		Address  string `json:"address"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if input.WeChatQR != "" {
		footerContact.WeChatQR = input.WeChatQR
	}
	if input.WeChatID != "" {
		footerContact.WeChatID = input.WeChatID
	}
	if input.Phone != "" {
		footerContact.Phone = input.Phone
	}
	if input.Email != "" {
		footerContact.Email = input.Email
	}
	if input.Address != "" {
		footerContact.Address = input.Address
	}

	if err := database.DB.Save(&footerContact).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update footer contact"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Footer contact updated successfully",
		"footer_contact": footerContact,
	})
}

// GetSystemLogs retrieves system logs with pagination and filtering
func GetSystemLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	userIDParam := c.Query("user_id")
	action := c.Query("action")
	objectType := c.Query("object_type")

	if database.UseMemory {
		allLogs := database.MemoryStore.GetAllLogs()
		var filtered []models.SystemLog
		for _, log := range allLogs {
			if userIDParam != "" {
				if uid, err := strconv.ParseUint(userIDParam, 10, 32); err == nil {
					if log.UserID != uint(uid) {
						continue
					}
				}
			}
			if action != "" && !strings.Contains(strings.ToLower(log.Action), strings.ToLower(action)) {
				continue
			}
			if objectType != "" && log.ObjectType != objectType {
				continue
			}
			filtered = append(filtered, log)
		}

		// Reverse order (newest first)
		sort.SliceStable(filtered, func(i, j int) bool {
			return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
		})

		total := int64(len(filtered))
		if offset >= len(filtered) {
			filtered = []models.SystemLog{}
		} else {
			end := offset + limit
			if end > len(filtered) {
				end = len(filtered)
			}
			filtered = filtered[offset:end]
		}

		c.JSON(http.StatusOK, gin.H{
			"logs": filtered,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		})
		return
	}

	query := database.DB.Model(&models.SystemLog{})

	// Apply filters
	if userIDParam != "" {
		if userID, err := strconv.ParseUint(userIDParam, 10, 32); err == nil {
			query = query.Where("user_id = ?", uint(userID))
		}
	}
	if action != "" {
		query = query.Where("action LIKE ?", "%"+action+"%")
	}
	if objectType != "" {
		query = query.Where("object_type = ?", objectType)
	}

	var logs []models.SystemLog
	var total int64

	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch system logs"})
		return
	}

	query.Model(&models.SystemLog{}).Count(&total)

	c.JSON(http.StatusOK, gin.H{
		"logs": logs,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// ServeDynamicPage serves a dynamic page based on navigation configuration
func ServeDynamicPage(c *gin.Context) {
	slug := c.Param("slug")

	if slug == "home" {
		c.File(filepath.Join("admin", "views", "index.html"))
		return
	}

	var navItem models.NavItem
	if database.UseMemory {
		item, err := database.MemoryStore.GetNavItemBySlug(slug)
		if err != nil {
			c.File(filepath.Join("admin", "views", slug+".html"))
			return
		}
		navItem = *item
	} else {
		if err := database.DB.Where("slug = ?", slug).First(&navItem).Error; err != nil {
			c.File(filepath.Join("admin", "views", slug+".html"))
			return
		}
	}

	switch navItem.NavType {
	case "blog":
		c.File(filepath.Join("admin", "views", "news-page.html"))
	case "service", "about":
		c.File(filepath.Join("admin", "views", "service-page.html"))
	default:
		c.File(filepath.Join("admin", "views", slug+".html"))
	}
}
