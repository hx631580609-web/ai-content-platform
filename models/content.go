package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// ContentType represents the type of content
type ContentType string

const (
	Article ContentType = "article"
	Poster  ContentType = "poster"
	Video   ContentType = "video"
)

// ContentInputType represents the source of content
type ContentInputType string

const (
	PromptInput    ContentInputType = "prompt"
	LinkInput      ContentInputType = "link"
	FileInput      ContentInputType = "file"
	PasteInput     ContentInputType = "paste"
)

// ContentBlockType represents the type of content block
type ContentBlockType string

const (
	TitleBlock    ContentBlockType = "title"
	ParagraphBlock ContentBlockType = "paragraph"
	ImageBlock    ContentBlockType = "image"
	ListBlock     ContentBlockType = "list"
	TableBlock    ContentBlockType = "table"
)

// Content represents the main content entity
type Content struct {
	ID          uint           `json:"id" gorm:"primary_key"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   *time.Time     `json:"deleted_at,omitempty" sql:"index"`
	UserID      uint           `json:"user_id" gorm:"not null"`
	Title       string         `json:"title" gorm:"not null"`
	Summary     string         `json:"summary"`
	Type        ContentType    `json:"type" gorm:"not null"`
	InputType   ContentInputType `json:"input_type" gorm:"not null"`
	Status      string         `json:"status" gorm:"default:'draft'"` // draft, published, archived
	ContentData string         `json:"content_data" gorm:"type:text"` // JSON string for structured content
	CoverImage  string         `json:"cover_image"`
	Tags             string         `json:"tags"` // comma-separated tags
	Category         string         `json:"category"` // 分类：沙特签证、中东商旅、出行指南、政策解读
	GeneratedContent string         `json:"generated_content" gorm:"type:text"` // AI生成的内容
	MetaTags         string         `json:"meta_tags"` // 元标签，如：沙特、商务签证、2026，逗号分隔，显示在内容卡片上
	SourceURL        string         `json:"source_url"` // for link input type

	// Relations
	User    User            `json:"user" gorm:"foreignkey:UserID"`
	Blocks  []ContentBlock  `json:"blocks" gorm:"foreignkey:ContentID"`
	Distributors []ContentDistribution `json:"distributions" gorm:"foreignkey:ContentID"`
}

// ContentBlock represents a structured content block
type ContentBlock struct {
	ID        uint             `json:"id" gorm:"primary_key"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	ContentID uint             `json:"content_id" gorm:"not null"`
	BlockType ContentBlockType `json:"block_type" gorm:"not null"`
	Content   string           `json:"content" gorm:"type:text"`
	MediaURL  string           `json:"media_url"` // for images, videos
	Order     int              `json:"order" gorm:"default:0"`
}

// ContentDistribution represents content distribution to various platforms
type ContentDistribution struct {
	ID          uint      `json:"id" gorm:"primary_key"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ContentID   uint      `json:"content_id" gorm:"not null"`
	Platform    string    `json:"platform"` // cms, wechat, video_channel, douyin, xiaohongshu
	Status      string    `json:"status" gorm:"default:'pending'"` // pending, processing, success, failed
	PublicationDate *time.Time `json:"publication_date"`
	URL         string    `json:"url"` // published URL
	ErrorMsg    string    `json:"error_msg"`
}

// TableName sets the table name for Content model
func (Content) TableName() string {
	return "contents"
}

// TableName sets the table name for ContentBlock model
func (ContentBlock) TableName() string {
	return "content_blocks"
}

// TableName sets the table name for ContentDistribution model
func (ContentDistribution) TableName() string {
	return "content_distributions"
}

// BeforeCreate hook for Content
func (c *Content) BeforeCreate(scope *gorm.Scope) error {
	if c.Status == "" {
		c.Status = "draft"
	}
	return nil
}