package admin

import (
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// ServeAdminInterface serves the admin interface pages
func ServeAdminInterface(c *gin.Context) {
	path := c.Request.URL.Path

	switch {
	// 登录页
	case path == "/login" || path == "/admin/login":
		c.File(filepath.Join("admin", "views", "login.html"))

	// 管理后台首页 / 仪表盘
	case path == "/" || path == "/admin" || path == "/admin/" ||
		path == "/admin/dashboard" || path == "/admin/dashboard/":
		c.File(filepath.Join("admin", "views", "admin_dashboard.html"))

	// 内容管理 - 列表
	case path == "/admin/contents" || path == "/admin/content_management":
		c.File(filepath.Join("admin", "views", "content_management.html"))

	// AI 创作助手 - HTML
	case path == "/admin/ai_assistant":
		c.File(filepath.Join("admin", "views", "ai_assistant.html"))
	// AI 创作助手 - JS
	case path == "/admin/ai-assistant.js":
		c.File(filepath.Join("admin", "views", "ai-assistant.js"))

	// 内容编辑器
	case path == "/admin/content_editor" || (len(path) > 22 && path[:22] == "/admin/content_editor?"):
		c.File(filepath.Join("admin", "views", "content_editor.html"))

	// 用户管理
	case path == "/admin/users" || path == "/admin/user_management":
		c.File(filepath.Join("admin", "views", "user_management.html"))

	// 博客管理
	case path == "/admin/blog-posts" || path == "/admin/blog_management":
		c.File(filepath.Join("admin", "views", "blog_management.html"))

	// 导航管理
	case path == "/admin/navigation" || path == "/admin/nav_management":
		c.File(filepath.Join("admin", "views", "nav_management.html"))

	// 博客编辑
	case path == "/admin/blog_edit" || (len(path) > 16 && path[:16] == "/admin/blog_edit?"):
		c.File(filepath.Join("admin", "views", "blog_edit.html"))

	// 服务页面管理
	case path == "/admin/service-pages" || path == "/admin/service_management":
		c.File(filepath.Join("admin", "views", "service_management.html"))

	// 系统日志
	case path == "/admin/system-logs" || path == "/admin/system_log_management":
		c.File(filepath.Join("admin", "views", "system_log_management.html"))

	// 官网前台页面
	case path == "/" || path == "/index" || path == "/home":
		c.File(filepath.Join("admin", "views", "website_homepage.html"))
	case path == "/visa/saudi":
		c.File(filepath.Join("admin", "views", "website_visa_saudi.html"))
	case path == "/visa/middle-east":
		c.File(filepath.Join("admin", "views", "website_visa_middle_east.html"))
	case path == "/transport":
		c.File(filepath.Join("admin", "views", "website_transport.html"))
	case path == "/business-expansion":
		c.File(filepath.Join("admin", "views", "website_business_expansion.html"))
	case path == "/business-trip":
		c.File(filepath.Join("admin", "views", "website_business_trip.html"))
	case path == "/insurance":
		c.File(filepath.Join("admin", "views", "website_insurance.html"))
	case path == "/about":
		c.File(filepath.Join("admin", "views", "website_about.html"))

	// 官网后台管理页面
	case path == "/admin/website/homepage":
		c.File(filepath.Join("admin", "views", "website_homepage_admin.html"))
	case path == "/admin/blog":
		c.File(filepath.Join("admin", "views", "blog_management.html"))
	case path == "/admin/business-pages":
		c.File(filepath.Join("admin", "views", "business_pages_admin.html"))
	case path == "/admin/about":
		c.File(filepath.Join("admin", "views", "website_about_admin.html"))
	case path == "/admin/footer":
		c.File(filepath.Join("admin", "views", "footer_admin.html"))
	case path == "/admin/logs":
		c.File(filepath.Join("admin", "views", "system_log_management.html"))

	// 首页模块编辑页
	case path == "/admin/edit_banner":
		c.File(filepath.Join("admin", "views", "edit_banner.html"))
	case path == "/admin/edit_trust_cards":
		c.File(filepath.Join("admin", "views", "edit_trust_cards.html"))
	case path == "/admin/edit_services_grid":
		c.File(filepath.Join("admin", "views", "edit_services_grid.html"))
	case path == "/admin/edit_blog_preview":
		c.File(filepath.Join("admin", "views", "edit_blog_preview.html"))
	case path == "/admin/edit_footer_contact":
		c.File(filepath.Join("admin", "views", "edit_footer_contact.html"))

	// 网站模块管理
	case path == "/admin/website/modules":
		c.File(filepath.Join("admin", "views", "website_module_management.html"))

	// 默认返回登录页
	default:
		c.File(filepath.Join("admin", "views", "login.html"))
	}
}
