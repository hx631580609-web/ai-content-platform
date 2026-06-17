package routes

import (
	"path/filepath"
	"strings"

	"ai-content-platform/admin"
	"ai-content-platform/handlers"
	"ai-content-platform/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(router *gin.Engine) {
	// Apply security headers globally
	router.Use(middleware.SecurityHeaders())

	// Serve static files for admin interface
	router.Static("/static", filepath.Join("admin", "assets"))

	// Serve images for website pages
	router.Static("/images", filepath.Join("admin", "views", "images"))

	// Serve CSS files
	router.Static("/css", filepath.Join("admin", "views", "css"))

	// Serve JS files
	router.Static("/js", filepath.Join("admin", "views", "js"))

	// Define API routes first to ensure they take precedence
	setupAPIRoutes(router)

	// Official website homepage route - 使用 zsts 参考项目的主页设计
	router.GET("/", func(c *gin.Context) {
		c.File(filepath.Join("admin", "views", "index.html"))
	})

	// Specific routes for admin interface pages
	router.GET("/login", func(c *gin.Context) {
		admin.ServeAdminInterface(c)
	})

	router.GET("/admin", func(c *gin.Context) {
		admin.ServeAdminInterface(c)
	})

	router.GET("/admin/*path", func(c *gin.Context) {
		// 统一由 admin_handler.go 处理所有 /admin/* 路径
		admin.ServeAdminInterface(c)
	})

	// 动态页面路由 - 根据导航配置自动路由到对应模板（放在最后，确保不会覆盖其他路由）
	router.GET("/:slug", func(c *gin.Context) {
		handlers.ServeDynamicPage(c)
	})

	// Admin interface routes - serve the admin dashboard for non-API routes
	router.NoRoute(func(c *gin.Context) {
		// Check if the path is for API routes that should be handled normally
		path := c.Request.URL.Path
		if isAPIRoute(path) {
			// This shouldn't happen as API routes are defined separately, but just in case
			c.Next()
			return
		}
		// For all other routes that aren't API routes, serve the admin interface
		admin.ServeAdminInterface(c)
	})
}

// setupAPIRoutes defines all the API routes
func setupAPIRoutes(router *gin.Engine) {
	// Public routes (no authentication required)
	router.POST("/register", handlers.Register)
	router.POST("/login", handlers.Login)

	// Public website routes
	router.GET("/website/modules", handlers.GetWebsiteModules)
	router.GET("/website/modules/:name", handlers.GetWebsiteModule)
	router.GET("/nav-items", handlers.GetNavItems)
	router.GET("/nav-items/:slug", handlers.GetNavItem)
	router.GET("/blog-posts", handlers.GetBlogPosts)
	router.GET("/blog-posts/:slug", handlers.GetBlogPost)
	router.GET("/service-pages", handlers.GetServicePages)
	router.GET("/service-pages/:slug", handlers.GetServicePage)
	router.GET("/about-us", handlers.GetAboutUs)
	router.GET("/footer-contact", handlers.GetFooterContact)

	// Navigation management API (no authentication required - admin panel internal operations)
	router.GET("/nav-items/all", handlers.GetAllNavItems)
	router.POST("/nav-items", handlers.CreateNavItem)
	router.PUT("/nav-items/:id", handlers.UpdateNavItem)
	router.DELETE("/nav-items/:id", handlers.DeleteNavItem)

	// Protected routes (authentication required)
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		// User routes
		protected.GET("/profile", handlers.GetProfile)
		protected.PUT("/profile", handlers.UpdateProfile)

		// Content routes
		protected.POST("/contents", middleware.InputValidation(), handlers.CreateContent)
		protected.GET("/contents", handlers.GetContents)
		protected.GET("/contents/statistics", handlers.GetContentStatistics)
		protected.GET("/contents/:id", handlers.GetContent)
		protected.PUT("/contents/:id", middleware.InputValidation(), handlers.UpdateContent)
		protected.DELETE("/contents/:id", handlers.DeleteContent)
		protected.POST("/contents/:id/publish", handlers.PublishContent)
		protected.POST("/contents/:id/archive", handlers.ArchiveContent)

		// AI content generation routes
		protected.POST("/ai/generate", handlers.GenerateContentWithAI)
		protected.POST("/ai/generate-blocks/:id", handlers.GenerateContentBlocks)

		// AI chat / LLM 对话接口（真实 LLM 调用）
		protected.GET("/ai/config", handlers.GetAIConfig)
		protected.POST("/ai/chat", handlers.ChatWithAI)
		protected.POST("/ai/chat/stream", handlers.ChatWithAIStream)

		// Blog and content management
		protected.POST("/blog-posts", middleware.InputValidation(), handlers.CreateBlogPost)
		protected.PUT("/blog-posts/:slug", middleware.InputValidation(), handlers.UpdateBlogPost)
		protected.DELETE("/blog-posts/:slug", handlers.DeleteBlogPost)
		protected.POST("/service-pages", middleware.InputValidation(), handlers.CreateServicePage)
		protected.PUT("/service-pages/:slug", middleware.InputValidation(), handlers.UpdateServicePage)
		protected.DELETE("/service-pages/:slug", handlers.DeleteServicePage)

		// Website settings
		protected.PUT("/website/modules/:name", middleware.InputValidation(), handlers.UpdateWebsiteModule)
		protected.PUT("/about-us", middleware.InputValidation(), handlers.UpdateAboutUs)
		protected.PUT("/footer-contact", middleware.InputValidation(), handlers.UpdateFooterContact)
	}

	// Admin routes (admin role required)
	adminGroup := router.Group("/")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		// User management
		adminGroup.GET("/users", handlers.GetAllUsers)
		adminGroup.GET("/users/:id", handlers.GetUserByID)
		adminGroup.PUT("/users/:id/role", handlers.UpdateUserRole)
		adminGroup.DELETE("/users/:id", handlers.DeleteUser)

		// System logs
		adminGroup.GET("/system-logs", handlers.GetSystemLogs)
	}
}

// isAPIRoute checks if the path is an API route that should be handled normally
func isAPIRoute(path string) bool {
	apiPaths := []string{
		"/register",
		"/login", // Only the POST /login is API, GET /login should show login page
		"/profile",
		"/contents",
		"/users",
		"/website/modules",
		"/blog-posts",
		"/service-pages",
		"/about-us",
		"/footer-contact",
		"/system-logs",
		"/ai",
	}

	for _, apiPath := range apiPaths {
		// Check if path matches API path but exclude GET /login which should show login page
		if (path == apiPath || path == apiPath+"/" ||
			strings.HasPrefix(path, apiPath+"/")) &&
			!(path == "/login" || path == "/login/") {
			return true
		}
	}
	return false
}
