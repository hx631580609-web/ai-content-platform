-- MySQL Database Initialization Script

-- Create the database if it doesn't exist
CREATE DATABASE IF NOT EXISTS ai_content_platform;

-- Use the database
USE ai_content_platform;

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role INT DEFAULT 1
);

-- Create contents table
CREATE TABLE IF NOT EXISTS contents (
    id INT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    user_id INT NOT NULL,
    title VARCHAR(255) NOT NULL,
    summary TEXT,
    type ENUM('article', 'poster', 'video') NOT NULL,
    input_type ENUM('prompt', 'link', 'file', 'paste') NOT NULL,
    status VARCHAR(50) DEFAULT 'draft',
    content_data LONGTEXT,
    cover_image VARCHAR(255),
    tags TEXT,
    source_url VARCHAR(255),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Create content_blocks table
CREATE TABLE IF NOT EXISTS content_blocks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    content_id INT NOT NULL,
    block_type ENUM('title', 'paragraph', 'image', 'list', 'table') NOT NULL,
    content LONGTEXT,
    media_url VARCHAR(255),
    `order` INT DEFAULT 0,
    FOREIGN KEY (content_id) REFERENCES contents(id) ON DELETE CASCADE
);

-- Create content_distributions table
CREATE TABLE IF NOT EXISTS content_distributions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    content_id INT NOT NULL,
    platform VARCHAR(100),
    status VARCHAR(50) DEFAULT 'pending',
    publication_date TIMESTAMP NULL,
    url VARCHAR(255),
    error_msg TEXT,
    FOREIGN KEY (content_id) REFERENCES contents(id) ON DELETE CASCADE
);

-- Create website_modules table
CREATE TABLE IF NOT EXISTS website_modules (
    id INT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    name VARCHAR(100) NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    position INT DEFAULT 0,
    config LONGTEXT
);

-- Create blog_posts table
CREATE TABLE IF NOT EXISTS blog_posts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    content LONGTEXT,
    summary TEXT,
    category VARCHAR(100),
    cover_image VARCHAR(255),
    published_at TIMESTAMP NULL,
    status VARCHAR(50) DEFAULT 'draft',
    view_count INT DEFAULT 0,
    author_id INT NOT NULL,
    FOREIGN KEY (author_id) REFERENCES users(id)
);

-- Create service_pages table
CREATE TABLE IF NOT EXISTS service_pages (
    id INT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    content LONGTEXT,
    service_type VARCHAR(100),
    cover_image VARCHAR(255),
    status VARCHAR(50) DEFAULT 'active',
    contact_info LONGTEXT,
    faq LONGTEXT
);

-- Create about_us table
CREATE TABLE IF NOT EXISTS about_us (
    id INT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    company_intro LONGTEXT,
    qualifications LONGTEXT,
    team_photos LONGTEXT,
    contact_info LONGTEXT
);

-- Create footer_contacts table
CREATE TABLE IF NOT EXISTS footer_contacts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    wechat_qr VARCHAR(255),
    wechat_id VARCHAR(100),
    phone VARCHAR(50),
    email VARCHAR(100),
    address TEXT
);

-- Create system_logs table
CREATE TABLE IF NOT EXISTS system_logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_id INT DEFAULT 0,
    action TEXT,
    object_type VARCHAR(100),
    object_id INT,
    description TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT
);

-- Insert default admin user (password is 'admin123' - hashed)
-- Note: This is a placeholder - in real implementation, you'd need to hash the password properly
INSERT INTO users (username, email, password, role) VALUES ('admin', 'admin@example.com', '$2a$14$NQzVKOo3PzdgfU5v2aBhked.ygYT2qOJ9.YF9NpKk.xkpwRb.V.Wu', 0) 
ON DUPLICATE KEY UPDATE username = username;

-- Insert default website modules
INSERT INTO website_modules (name, enabled, position, config) VALUES 
('banner', true, 1, '{"title": "沙特商务签证·官方授权代办", "subtitle": "专业、高效、可靠的签证服务", "cta_text": "立即咨询", "background_image": "/images/banner-bg.jpg"}'),
('trust_cards', true, 2, '{"data": [{"label": "办理国家", "value": "15+"}, {"label": "出签率", "value": "98%"}, {"label": "服务客户", "value": "5000+"}]}'),
('services_grid', true, 3, '{"services": [{"title": "沙特商务签证", "desc": "官方授权，快速办理"}, {"title": "签证业务", "desc": "中东北非各国签证"}, {"title": "交通服务", "desc": "接送机、租车服务"}, {"title": "企业出海", "desc": "海外市场拓展"}, {"title": "企业考察", "desc": "商务考察安排"}]}'),
('blog_preview', true, 4, '{"count": 3}'),
('footer_contact', true, 5, '{"show_wechat": true, "show_phone": true, "show_email": true}')
ON DUPLICATE KEY UPDATE name = name;

-- Insert default about us content
INSERT INTO about_us (company_intro, qualifications, team_photos, contact_info) VALUES 
('中盛启瀚是一家专业的沙特商务签证代办服务机构，致力于为中国企业及个人提供高效、可靠的签证服务。', '{"licenses": ["官方授权证书", "ISO认证", "行业协会会员"]}', '["/images/team1.jpg", "/images/team2.jpg"]', '{"phone": "400-XXX-XXXX", "email": "info@heshengvisa.com", "address": "北京市朝阳区XXX大厦"}')
ON DUPLICATE KEY UPDATE company_intro = company_intro;

-- Insert default footer contact
INSERT INTO footer_contacts (wechat_qr, wechat_id, phone, email, address) VALUES 
('/images/wechat-qr.jpg', 'heshengvisa', '400-XXX-XXXX', 'info@heshengvisa.com', '北京市朝阳区XXX大厦')
ON DUPLICATE KEY UPDATE wechat_id = wechat_id;