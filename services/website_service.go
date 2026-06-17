package services

import (
	"errors"
	"fmt"

	"ai-content-platform/database"
	"ai-content-platform/models"

	"github.com/jinzhu/gorm"
)

// WebsiteService provides methods for managing website content
type WebsiteService struct{}

// GetActiveBlogPosts retrieves published blog posts with optional category filter
func (ws *WebsiteService) GetActiveBlogPosts(category *string, limit, offset int) ([]models.BlogPost, int, error) {
	query := database.DB.Preload("Author", "id, username, email").Model(&models.BlogPost{}).Where("status = ?", "published")

	if category != nil {
		query = query.Where("category = ?", *category)
	}

	var total int
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var blogPosts []models.BlogPost
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&blogPosts).Error; err != nil {
		return nil, 0, err
	}

	return blogPosts, total, nil
}

// GetBlogPostBySlug retrieves a published blog post by its slug
func (ws *WebsiteService) GetBlogPostBySlug(slug string) (*models.BlogPost, error) {
	var blogPost models.BlogPost
	err := database.DB.Preload("Author", "id, username, email").First(&blogPost, "slug = ? AND status = ?", slug, "published").Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.New("blog post not found")
		}
		return nil, err
	}

	// Increment view count
	blogPost.ViewCount++
	database.DB.Save(&blogPost)

	return &blogPost, nil
}

// GetServicePages retrieves service pages with optional filters
func (ws *WebsiteService) GetServicePages(serviceType *string, status *string, limit, offset int) ([]models.ServicePage, int, error) {
	query := database.DB.Model(&models.ServicePage{})

	if serviceType != nil {
		query = query.Where("service_type = ?", *serviceType)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	var total int
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var servicePages []models.ServicePage
	if err := query.Offset(offset).Limit(limit).Find(&servicePages).Error; err != nil {
		return nil, 0, err
	}

	return servicePages, total, nil
}

// GetServicePageBySlug retrieves a service page by its slug
func (ws *WebsiteService) GetServicePageBySlug(slug string) (*models.ServicePage, error) {
	var servicePage models.ServicePage
	err := database.DB.First(&servicePage, "slug = ?", slug).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.New("service page not found")
		}
		return nil, err
	}

	return &servicePage, nil
}

// GetWebsiteModules retrieves all enabled website modules ordered by position
func (ws *WebsiteService) GetWebsiteModules() ([]models.WebsiteModule, error) {
	var modules []models.WebsiteModule
	err := database.DB.Where("enabled = ?", true).Order("position ASC").Find(&modules).Error
	if err != nil {
		return nil, err
	}

	return modules, nil
}

// GetWebsiteModuleByName retrieves a specific website module by name
func (ws *WebsiteService) GetWebsiteModuleByName(name string) (*models.WebsiteModule, error) {
	var module models.WebsiteModule
	err := database.DB.First(&module, "name = ?", name).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.New("website module not found")
		}
		return nil, err
	}

	return &module, nil
}

// UpdateWebsiteModule updates a website module
func (ws *WebsiteService) UpdateWebsiteModule(name string, enabled *bool, position *int, config *string) error {
	var module models.WebsiteModule
	err := database.DB.First(&module, "name = ?", name).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("website module not found")
		}
		return err
	}

	if enabled != nil {
		module.Enabled = *enabled
	}
	if position != nil {
		module.Position = *position
	}
	if config != nil {
		module.Config = *config
	}

	return database.DB.Save(&module).Error
}

// GetAboutUs retrieves the about us page content
func (ws *WebsiteService) GetAboutUs() (*models.AboutUs, error) {
	var aboutUs models.AboutUs
	err := database.DB.First(&aboutUs).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			// Return default empty content if not found
			return &models.AboutUs{}, nil
		}
		return nil, err
	}

	return &aboutUs, nil
}

// UpdateAboutUs updates the about us page content
func (ws *WebsiteService) UpdateAboutUs(companyIntro, qualifications, teamPhotos, contactInfo *string) error {
	var aboutUs models.AboutUs

	// Try to find existing record, if not found, create a new one
	err := database.DB.First(&aboutUs).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			// Create new record
			aboutUs = models.AboutUs{}
		} else {
			return err
		}
	}

	if companyIntro != nil {
		aboutUs.CompanyIntro = *companyIntro
	}
	if qualifications != nil {
		aboutUs.Qualifications = *qualifications
	}
	if teamPhotos != nil {
		aboutUs.TeamPhotos = *teamPhotos
	}
	if contactInfo != nil {
		aboutUs.ContactInfo = *contactInfo
	}

	return database.DB.Save(&aboutUs).Error
}

// GetFooterContact retrieves the footer contact information
func (ws *WebsiteService) GetFooterContact() (*models.FooterContact, error) {
	var footerContact models.FooterContact
	err := database.DB.First(&footerContact).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			// Return default empty content if not found
			return &models.FooterContact{}, nil
		}
		return nil, err
	}

	return &footerContact, nil
}

// UpdateFooterContact updates the footer contact information
func (ws *WebsiteService) UpdateFooterContact(weChatQR, weChatID, phone, email, address *string) error {
	var footerContact models.FooterContact

	// Try to find existing record, if not found, create a new one
	err := database.DB.First(&footerContact).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			// Create new record
			footerContact = models.FooterContact{}
		} else {
			return err
		}
	}

	if weChatQR != nil {
		footerContact.WeChatQR = *weChatQR
	}
	if weChatID != nil {
		footerContact.WeChatID = *weChatID
	}
	if phone != nil {
		footerContact.Phone = *phone
	}
	if email != nil {
		footerContact.Email = *email
	}
	if address != nil {
		footerContact.Address = *address
	}

	return database.DB.Save(&footerContact).Error
}

// CreateBlogPost creates a new blog post
func (ws *WebsiteService) CreateBlogPost(title, slug, content, summary, category, coverImage string, authorID uint, status string) (*models.BlogPost, error) {
	// Check if slug already exists
	var existingPost models.BlogPost
	if database.DB.Where("slug = ?", slug).First(&existingPost).Error == nil {
		return nil, errors.New("blog post with this slug already exists")
	}

	blogPost := models.BlogPost{
		Title:      title,
		Slug:       slug,
		Content:    content,
		Summary:    summary,
		Category:   category,
		CoverImage: coverImage,
		AuthorID:   authorID,
		Status:     status,
	}

	// Set default status to draft if not provided
	if blogPost.Status == "" {
		blogPost.Status = "draft"
	}

	if err := database.DB.Create(&blogPost).Error; err != nil {
		return nil, err
	}

	return &blogPost, nil
}

// UpdateBlogPost updates an existing blog post
func (ws *WebsiteService) UpdateBlogPost(slug, title, newSlug, content, summary, category, coverImage, status string) error {
	var blogPost models.BlogPost
	err := database.DB.First(&blogPost, "slug = ?", slug).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("blog post not found")
		}
		return err
	}

	// If new slug is provided, check if it already exists
	if newSlug != "" && newSlug != slug {
		var existingPost models.BlogPost
		if database.DB.Where("slug = ? AND id != ?", newSlug, blogPost.ID).First(&existingPost).Error == nil {
			return errors.New("blog post with this slug already exists")
		}
		blogPost.Slug = newSlug
	}

	// Update fields if provided
	if title != "" {
		blogPost.Title = title
	}
	if content != "" {
		blogPost.Content = content
	}
	if summary != "" {
		blogPost.Summary = summary
	}
	if category != "" {
		blogPost.Category = category
	}
	if coverImage != "" {
		blogPost.CoverImage = coverImage
	}
	if status != "" {
		blogPost.Status = status
	}

	return database.DB.Save(&blogPost).Error
}

// DeleteBlogPost deletes a blog post by slug
func (ws *WebsiteService) DeleteBlogPost(slug string) error {
	var blogPost models.BlogPost
	err := database.DB.First(&blogPost, "slug = ?", slug).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("blog post not found")
		}
		return err
	}

	return database.DB.Delete(&blogPost).Error
}

// CreateServicePage creates a new service page
func (ws *WebsiteService) CreateServicePage(title, slug, description, content, serviceType, coverImage, contactInfo, faq string, status string) (*models.ServicePage, error) {
	// Check if slug already exists
	var existingPage models.ServicePage
	if database.DB.Where("slug = ?", slug).First(&existingPage).Error == nil {
		return nil, errors.New("service page with this slug already exists")
	}

	servicePage := models.ServicePage{
		Title:       title,
		Slug:        slug,
		Description: description,
		Content:     content,
		ServiceType: serviceType,
		CoverImage:  coverImage,
		Status:      status,
		ContactInfo: contactInfo,
		Faq:         faq,
	}

	// Set default status to active if not provided
	if servicePage.Status == "" {
		servicePage.Status = "active"
	}

	if err := database.DB.Create(&servicePage).Error; err != nil {
		return nil, err
	}

	return &servicePage, nil
}

// UpdateServicePage updates an existing service page
func (ws *WebsiteService) UpdateServicePage(slug, title, newSlug, description, content, serviceType, coverImage, status, contactInfo, faq string) error {
	var servicePage models.ServicePage
	err := database.DB.First(&servicePage, "slug = ?", slug).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("service page not found")
		}
		return err
	}

	// If new slug is provided, check if it already exists
	if newSlug != "" && newSlug != slug {
		var existingPage models.ServicePage
		if database.DB.Where("slug = ? AND id != ?", newSlug, servicePage.ID).First(&existingPage).Error == nil {
			return errors.New("service page with this slug already exists")
		}
		servicePage.Slug = newSlug
	}

	// Update fields if provided
	if title != "" {
		servicePage.Title = title
	}
	if description != "" {
		servicePage.Description = description
	}
	if content != "" {
		servicePage.Content = content
	}
	if serviceType != "" {
		servicePage.ServiceType = serviceType
	}
	if coverImage != "" {
		servicePage.CoverImage = coverImage
	}
	if status != "" {
		servicePage.Status = status
	}
	if contactInfo != "" {
		servicePage.ContactInfo = contactInfo
	}
	if faq != "" {
		servicePage.Faq = faq
	}

	return database.DB.Save(&servicePage).Error
}

// DeleteServicePage deletes a service page by slug
func (ws *WebsiteService) DeleteServicePage(slug string) error {
	var servicePage models.ServicePage
	err := database.DB.First(&servicePage, "slug = ?", slug).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("service page not found")
		}
		return err
	}

	return database.DB.Delete(&servicePage).Error
}

// GetRecentBlogPosts retrieves the most recent blog posts
func (ws *WebsiteService) GetRecentBlogPosts(count int) ([]models.BlogPost, error) {
	var blogPosts []models.BlogPost
	err := database.DB.Preload("Author", "id, username, email").
		Where("status = ?", "published").
		Order("created_at DESC").
		Limit(count).
		Find(&blogPosts).Error

	if err != nil {
		return nil, err
	}

	return blogPosts, nil
}

// GetSystemLogs retrieves system logs with optional filters
func (ws *WebsiteService) GetSystemLogs(userID *uint, action, objectType string, limit, offset int) ([]models.SystemLog, int, error) {
	query := database.DB.Model(&models.SystemLog{})

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if action != "" {
		query = query.Where("action LIKE ?", fmt.Sprintf("%%%s%%", action))
	}
	if objectType != "" {
		query = query.Where("object_type = ?", objectType)
	}

	var total int
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []models.SystemLog
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
