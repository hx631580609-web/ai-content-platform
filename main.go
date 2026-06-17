package main

import (
	"log"

	"ai-content-platform/config"
	"ai-content-platform/database"
	"ai-content-platform/models"
	"ai-content-platform/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	log.Printf("Starting server on port %s", cfg.ServerPort)

	// Initialize database
	log.Println("Connecting to database...")
	database.ConnectDatabase()
	defer database.CloseDatabase()

	// Check if database connection is valid before proceeding
	if database.DB == nil {
		log.Println("Warning: Database connection failed, some features will be unavailable")
	} else {
		log.Println("Database connected successfully")

		// Auto migrate the schema
		log.Println("Auto migrating schema...")
		database.DB.AutoMigrate(
			&models.User{},
			&models.Content{},
			&models.ContentBlock{},
			&models.ContentDistribution{},
			&models.WebsiteModule{},
			&models.NavItem{},
			&models.BlogPost{},
			&models.ServicePage{},
			&models.AboutUs{},
			&models.FooterContact{},
			&models.SystemLog{},
		)
		log.Println("Schema migration completed")

		// Create default admin user if not exists
		createDefaultAdmin()

		// Create default website modules if not exist
		createDefaultWebsiteModules()

		// Create default about us and footer contact if not exist
		createDefaultWebsiteContent()

		// Create default navigation items if not exist
		createDefaultNavItems()

		// Create default service pages if not exist
		createDefaultServicePages()
	}

	// Initialize Gin engine
	gin.SetMode(gin.DebugMode) // Changed to debug mode for better logging
	router := gin.New()

	// Global middleware - minimal for debugging
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Setup routes
	log.Println("Setting up routes...")
	routes.SetupRoutes(router)
	log.Println("Routes setup completed")

	// Start the server
	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// createDefaultAdmin creates a default admin user if no users exist
func createDefaultAdmin() {
	if database.DB == nil {
		log.Println("Database not available, skipping default admin creation")
		return
	}

	var count int
	database.DB.Model(&models.User{}).Count(&count)

	if count == 0 {
		adminUser := models.User{
			Username: "admin",
			Email:    "admin@example.com",
			Password: "admin123", // This will be hashed by the BeforeCreate hook
			Role:     models.Admin,
		}

		if err := database.DB.Create(&adminUser).Error; err != nil {
			log.Printf("Failed to create default admin user: %v", err)
		} else {
			log.Println("Default admin user created: admin / admin123")
		}
	} else {
		log.Printf("Found %d existing users, skipping default admin creation", count)
	}
}

// createDefaultWebsiteModules creates default website modules if none exist
func createDefaultWebsiteModules() {
	if database.DB == nil {
		log.Println("Database not available, skipping default website modules creation")
		return
	}

	var count int
	database.DB.Model(&models.WebsiteModule{}).Count(&count)

	if count == 0 {
		modules := []models.WebsiteModule{
			{
				Name:     "banner",
				Enabled:  true,
				Position: 1,
				Config:   `{"title": "沙特商务签证·官方授权代办", "subtitle": "专业、高效、可靠的签证服务", "cta_text": "立即咨询", "background_image": "/images/banner-bg.jpg"}`,
			},
			{
				Name:     "trust_cards",
				Enabled:  true,
				Position: 2,
				Config:   `{"data": [{"label": "办理国家", "value": "15+"}, {"label": "出签�?, "value": "98%"}, {"label": "服务客户", "value": "5000+"}]}`,
			},
			{
				Name:     "services_grid",
				Enabled:  true,
				Position: 3,
				Config:   `{"services": [{"title": "沙特商务签证", "desc": "官方授权，快速办�?}, {"title": "签证业务", "desc": "中东北非各国签证"}, {"title": "交通服�?, "desc": "接送机、租车服�?}, {"title": "企业出海", "desc": "海外市场拓展"}, {"title": "企业考察", "desc": "商务考察安排"}]}`,
			},
			{
				Name:     "blog_preview",
				Enabled:  true,
				Position: 4,
				Config:   `{"count": 3}`,
			},
			{
				Name:     "footer_contact",
				Enabled:  true,
				Position: 5,
				Config:   `{"show_wechat": true, "show_phone": true, "show_email": true}`,
			},
		}

		for _, module := range modules {
			if err := database.DB.Create(&module).Error; err != nil {
				log.Printf("Failed to create default website module %s: %v", module.Name, err)
			}
		}

		log.Println("Default website modules created")
	} else {
		log.Printf("Found %d existing website modules, skipping default creation", count)
	}
}

// createDefaultWebsiteContent creates default about us and footer contact if none exist
func createDefaultWebsiteContent() {
	if database.DB == nil {
		log.Println("Database not available, skipping default website content creation")
		return
	}

	// Check if about us exists
	var aboutUsCount int
	database.DB.Model(&models.AboutUs{}).Count(&aboutUsCount)

	if aboutUsCount == 0 {
		aboutUs := models.AboutUs{
			CompanyIntro:   "中盛启瀚是一家专业的沙特商务签证代办服务机构，致力于为中国企业及个人提供高效、可靠的签证服务",
			Qualifications: `{"licenses": ["官方授权证书", "ISO 认证", "行业协会会员"]}`,
			TeamPhotos:     `["/images/team1.jpg", "/images/team2.jpg"]`,
			ContactInfo:    `{"phone": "400-XXX-XXXX", "email": "info@heshengvisa.com", "address": "北京市朝阳区 XXX 大厦"}`,
		}

		if err := database.DB.Create(&aboutUs).Error; err != nil {
			log.Printf("Failed to create default about us content: %v", err)
		} else {
			log.Println("Default about us content created")
		}
	} else {
		log.Printf("Found existing about us content, skipping default creation (%d)", aboutUsCount)
	}

	// Check if footer contact exists
	var footerContactCount int
	database.DB.Model(&models.FooterContact{}).Count(&footerContactCount)

	if footerContactCount == 0 {
		footerContact := models.FooterContact{
			WeChatQR: "/images/wechat-qr.jpg",
			WeChatID: "heshengvisa",
			Phone:    "400-XXX-XXXX",
			Email:    "info@heshengvisa.com",
			Address:  "北京市朝阳区 XXX 大厦",
		}

		if err := database.DB.Create(&footerContact).Error; err != nil {
			log.Printf("Failed to create default footer contact: %v", err)
		} else {
			log.Println("Default footer contact created")
		}
	} else {
		log.Printf("Found existing footer contact, skipping default creation (%d)", footerContactCount)
	}
}

// createDefaultNavItems creates default navigation items if none exist
func createDefaultNavItems() {
	if database.DB == nil {
		log.Println("Database not available, skipping default nav items creation")
		return
	}

	var count int
	database.DB.Model(&models.NavItem{}).Count(&count)

	if count == 0 {
		navItems := []models.NavItem{
			{Name: "沙特签证", NameEn: "Saudi Visa", Slug: "saudi-visa", NavType: "service", LinkType: "internal", ServiceType: "saudi_business_visa", Position: 1, Enabled: true},
			{Name: "全球签证", NameEn: "Global Visa", Slug: "visa", NavType: "service", LinkType: "internal", ServiceType: "other_destinations_visa", Position: 2, Enabled: true},
			{Name: "境外交通住宿", NameEn: "Transport & Accommodation", Slug: "transport", NavType: "service", LinkType: "internal", ServiceType: "transport", Position: 3, Enabled: true},
			{Name: "境外保险", NameEn: "Insurance", Slug: "insurance", NavType: "service", LinkType: "internal", ServiceType: "insurance", Position: 4, Enabled: true},
			{Name: "企业出海", NameEn: "Enterprise Outbound", Slug: "enterprise", NavType: "service", LinkType: "internal", ServiceType: "enterprise_outbound", Position: 5, Enabled: true},
			{Name: "企业考察", NameEn: "Enterprise Inspection", Slug: "inspection", NavType: "service", LinkType: "internal", ServiceType: "enterprise_inspection", Position: 6, Enabled: true},
			{Name: "沙特资讯", NameEn: "Saudi News", Slug: "news", NavType: "blog", LinkType: "internal", Position: 7, Enabled: true},
			{Name: "关于我们", NameEn: "About Us", Slug: "about", NavType: "about", LinkType: "internal", Position: 8, Enabled: true},
		}

		for _, item := range navItems {
			if err := database.DB.Create(&item).Error; err != nil {
				log.Printf("Failed to create default nav item %s: %v", item.Name, err)
			}
		}

		log.Println("Default navigation items created")
	} else {
		log.Printf("Found %d existing navigation items, skipping default creation", count)
	}
}

// createDefaultServicePages creates or updates default service pages with rich content
func createDefaultServicePages() {
	if database.DB == nil {
		log.Println("Database not available, skipping default service pages creation")
		return
	}

	// Define service pages data with rich content
	servicePagesData := []struct {
		Title, Slug, Description, Content, ServiceType, CoverImage string
	}{
		{
			Title:       "沙特签证服务",
			Slug:        "saudi-visa",
			Description: "专业、高效的沙特阿拉伯签证办理服务，涵盖商务签证、工作签证、旅游签证等多种类型",
			Content:     "<div class='max-w-6xl mx-auto'><h2 class='text-3xl font-bold mb-6 text-gray-900'>沙特签证服务</h2><p class='text-lg text-gray-600 mb-8'>我们提供专业的沙特阿拉伯签证办理服务，包括商务签证、工作签证、旅游签证等多种类型</p></div>",
			ServiceType: "saudi_business_visa",
			CoverImage:  "/images/riyadh-skyline.jpg",
		},
		{
			Title:       "全球签证服务",
			Slug:        "visa",
			Description: "覆盖全球100+国家的签证办理服务，多国签证一站搞定",
			Content:     "<div class='max-w-6xl mx-auto'><h2 class='text-3xl font-bold mb-6 text-gray-900'>全球签证服务</h2><p class='text-lg text-gray-600 mb-8'>我们与全球各国领事馆保持良好关系，为您提供便捷的签证办理服务，覆盖100多个国家和地区</p></div>",
			ServiceType: "other_destinations_visa",
			CoverImage:  "/images/global-visa.jpg",
		},
		{
			Title:       "境外交通住宿",
			Slug:        "transport",
			Description: "一站式境外交通与住宿预订服务，让您的出行更加便捷",
			Content:     "<div class='max-w-6xl mx-auto'><h2 class='text-3xl font-bold mb-6 text-gray-900'>境外交通住宿</h2><p class='text-lg text-gray-600 mb-8'>我们提供境外机票、酒店、专车等一站式预订服务，让您的出行更加便捷舒心</p></div>",
			ServiceType: "transport",
			CoverImage:  "/images/transport.jpg",
		},
		{
			Title:       "境外保险服务",
			Slug:        "insurance",
			Description: "全面的境外旅行保险与健康保障，让您的旅途更加安心",
			Content:     "<div class='max-w-6xl mx-auto'><h2 class='text-3xl font-bold mb-6 text-gray-900'>境外保险服务</h2><p class='text-lg text-gray-600 mb-8'>我们提供多种境外保险产品，包括旅行险、医疗险、意外险等，全方位保障您的出行安全</p></div>",
			ServiceType: "insurance",
			CoverImage:  "/images/insurance.jpg",
		},
		{
			Title:       "企业出海服务",
			Slug:        "enterprise",
			Description: "全方位企业海外拓展解决方案，助力企业顺利出海",
			Content:     "<div class='max-w-6xl mx-auto'><h2 class='text-3xl font-bold mb-6 text-gray-900'>企业出海服务</h2><p class='text-lg text-gray-600 mb-8'>我们为企业提供海外市场调研、公司注册、法律咨询等一站式服务，助力企业顺利出海，拓展全球市场</p></div>",
			ServiceType: "enterprise_outbound",
			CoverImage:  "/images/enterprise.jpg",
		},
		{
			Title:       "企业考察服务",
			Slug:        "inspection",
			Description: "专业的海外商务考察与项目对接，帮助企业发现商机",
			Content:     "<div class='max-w-6xl mx-auto'><h2 class='text-3xl font-bold mb-6 text-gray-900'>企业考察服务</h2><p class='text-lg text-gray-600 mb-8'>我们为企业提供海外商务考察、项目对接、实地调研等专业服务，帮助企业发现商机、拓展合作</p></div>",
			ServiceType: "enterprise_inspection",
			CoverImage:  "/images/inspection.jpg",
		},
	}

	// Upsert each service page
	for _, sp := range servicePagesData {
		var existing models.ServicePage
		result := database.DB.Where("slug = ?", sp.Slug).First(&existing)

		if result.Error != nil {
			// Create new if not exists
			page := models.ServicePage{
				Title:       sp.Title,
				Slug:        sp.Slug,
				Description: sp.Description,
				Content:     sp.Content,
				ServiceType: sp.ServiceType,
				CoverImage:  sp.CoverImage,
				Status:      "active",
			}
			if err := database.DB.Create(&page).Error; err != nil {
				log.Printf("Failed to create service page %s: %v", sp.Slug, err)
			} else {
				log.Printf("Created service page: %s", sp.Title)
			}
		} else {
			// Update existing
			existing.Title = sp.Title
			existing.Description = sp.Description
			existing.Content = sp.Content
			existing.ServiceType = sp.ServiceType
			existing.CoverImage = sp.CoverImage
			if err := database.DB.Save(&existing).Error; err != nil {
				log.Printf("Failed to update service page %s: %v", sp.Slug, err)
			} else {
				log.Printf("Updated service page: %s", sp.Title)
			}
		}
	}

	log.Println("Service pages initialization completed")
}
