package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// NavItem represents a navigation item in the website header
type NavItem struct {
	ID          uint      `json:"id" gorm:"primary_key"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `json:"name" gorm:"not null"`
	NameEn      string    `json:"name_en"`
	Slug        string    `json:"slug" gorm:"unique_index;not null"`
	NavType     string    `json:"nav_type" gorm:"not null"` // service, blog, about, external
	LinkType    string    `json:"link_type"`                // internal, external
	URL         string    `json:"url"`                      // for external links
	ServiceType string    `json:"service_type"`             // for service type nav items
	Position    int       `json:"position" gorm:"default:0"`
	Enabled     bool      `json:"enabled" gorm:"default:true"`
	ParentID    uint      `json:"parent_id"` // for dropdown items
}

// TableName sets the table name for NavItem model
func (NavItem) TableName() string {
	return "nav_items"
}

// WebsiteModule represents different modules on the website homepage
type WebsiteModule struct {
	ID        uint      `json:"id" gorm:"primary_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name" gorm:"not null"` // banner, trust_cards, services_grid, blog_preview, footer_contact
	Enabled   bool      `json:"enabled" gorm:"default:true"`
	Position  int       `json:"position" gorm:"default:0"`
	Config    string    `json:"config" gorm:"type:text"` // JSON string for module-specific config
}

// BlogPost represents a blog post for the Saudi Arabia information site
type BlogPost struct {
	ID            uint       `json:"id" gorm:"primary_key"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Title         string     `json:"title" gorm:"not null"`
	Slug          string     `json:"slug" gorm:"unique_index;not null"`
	Content       string     `json:"content" gorm:"type:text"`
	Summary       string     `json:"summary"`
	Category      string     `json:"category"` // saudi_visa, middle_east_travel, travel_guide, policy_interpretation
	CoverImage    string     `json:"cover_image"`
	PublishedAt   *time.Time `json:"published_at"`
	Status        string     `json:"status" gorm:"default:'draft'"` // draft, published, archived
	ViewCount     int        `json:"view_count" gorm:"default:0"`
	AuthorID      uint       `json:"author_id" gorm:"not null"`
	IsAiGenerated bool       `json:"is_ai_generated" gorm:"default:false"`

	Author User `json:"author" gorm:"foreignkey:AuthorID"`
}

// ServicePage represents business service pages
type ServicePage struct {
	ID          uint      `json:"id" gorm:"primary_key"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Title       string    `json:"title" gorm:"not null"`
	Slug        string    `json:"slug" gorm:"unique_index;not null"`
	Description string    `json:"description"`
	Content     string    `json:"content" gorm:"type:text"`
	ServiceType string    `json:"service_type"` // saudi_business_visa, other_destinations_visa, transport, enterprise_outbound, enterprise_inspection, insurance
	CoverImage  string    `json:"cover_image"`
	Status      string    `json:"status" gorm:"default:'active'"` // active, inactive
	ContactInfo string    `json:"contact_info" gorm:"type:text"`  // JSON string for contact information
	Faq         string    `json:"faq" gorm:"type:text"`           // JSON string for frequently asked questions
}

// AboutUs represents the about us page content
type AboutUs struct {
	ID             uint      `json:"id" gorm:"primary_key"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	CompanyIntro   string    `json:"company_intro" gorm:"type:text"`
	Qualifications string    `json:"qualifications" gorm:"type:text"` // JSON string for qualifications and certifications
	TeamPhotos     string    `json:"team_photos" gorm:"type:text"`    // JSON string for team photos
	ContactInfo    string    `json:"contact_info" gorm:"type:text"`   // JSON string for contact information
}

// FooterContact represents footer contact information
type FooterContact struct {
	ID        uint      `json:"id" gorm:"primary_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	WeChatQR  string    `json:"wechat_qr"` // WeChat QR code image URL
	WeChatID  string    `json:"wechat_id"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Address   string    `json:"address"`
}

// SystemLog represents system logs for audit trail
type SystemLog struct {
	ID          uint      `json:"id" gorm:"primary_key"`
	CreatedAt   time.Time `json:"created_at"`
	UserID      uint      `json:"user_id"`                // 0 for system operations
	Action      string    `json:"action" gorm:"not null"` // create, edit, delete, publish, etc.
	ObjectType  string    `json:"object_type"`            // user, content, blog_post, service_page, etc.
	ObjectID    uint      `json:"object_id"`
	Description string    `json:"description"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
}

// TableName sets the table name for WebsiteModule model
func (WebsiteModule) TableName() string {
	return "website_modules"
}

// TableName sets the table name for BlogPost model
func (BlogPost) TableName() string {
	return "blog_posts"
}

// TableName sets the table name for ServicePage model
func (ServicePage) TableName() string {
	return "service_pages"
}

// TableName sets the table name for AboutUs model
func (AboutUs) TableName() string {
	return "about_us"
}

// TableName sets the table name for FooterContact model
func (FooterContact) TableName() string {
	return "footer_contacts"
}

// TableName sets the table name for SystemLog model
func (SystemLog) TableName() string {
	return "system_logs"
}

// BeforeCreate hook for BlogPost
func (bp *BlogPost) BeforeCreate(scope *gorm.Scope) error {
	if bp.Status == "" {
		bp.Status = "draft"
	}
	return nil
}

// BeforeCreate hook for ServicePage
func (sp *ServicePage) BeforeCreate(scope *gorm.Scope) error {
	if sp.Status == "" {
		sp.Status = "active"
	}
	return nil
}
