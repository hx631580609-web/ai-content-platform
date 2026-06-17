package database

import (
	"errors"
	"sync"
	"time"

	"ai-content-platform/models"
)

var (
	MemoryStore *InMemoryStore
	UseMemory   bool
)

type InMemoryStore struct {
	mu            sync.RWMutex
	users         map[uint]*models.User
	usernameIndex map[string]uint
	userCounter   uint

	contents       map[uint]*models.Content
	contentIndex   map[string][]uint
	contentCounter uint

	blocks       map[uint]*models.ContentBlock
	blockCounter uint

	blogs       map[uint]*models.BlogPost
	blogSlugMap map[string]uint
	blogCounter uint

	websiteModules map[string]map[string]interface{}

	servicePages   map[string]*models.ServicePage
	serviceCounter uint

	navItems   map[uint]*models.NavItem
	navSlugMap map[string]uint
	navCounter uint

	logs       []models.SystemLog
	logCounter uint
}

func NewInMemoryStore() *InMemoryStore {
	store := &InMemoryStore{
		users:         make(map[uint]*models.User),
		usernameIndex: make(map[string]uint),
		userCounter:   0,

		contents:     make(map[uint]*models.Content),
		contentIndex: make(map[string][]uint),

		blocks:       make(map[uint]*models.ContentBlock),
		blockCounter: 0,

		blogs:       make(map[uint]*models.BlogPost),
		blogSlugMap: make(map[string]uint),
		blogCounter: 0,

		websiteModules: make(map[string]map[string]interface{}),

		servicePages: make(map[string]*models.ServicePage),

		navItems:   make(map[uint]*models.NavItem),
		navSlugMap: make(map[string]uint),

		logs:       []models.SystemLog{},
		logCounter: 0,
	}

	// Create default admin user
	adminUser := &models.User{
		ID:        1,
		Username:  "admin",
		Email:     "admin@example.com",
		Password:  "admin123",
		Role:      models.Admin,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	store.users[1] = adminUser
	store.usernameIndex["admin"] = 1
	store.userCounter = 1

	// Create some default website modules
	store.websiteModules["banner"] = map[string]interface{}{
		"title":       "沙特商务签证",
		"subtitle":    "官方授权代办",
		"description": "专业办理沙特商务签证及中东地区各类商旅服务",
	}

	store.websiteModules["trust_cards"] = map[string]interface{}{
		"countries":     15,
		"approval_rate": 98,
		"clients":       5000,
		"experience":    15,
	}

	// Create default navigation items
	defaultNavItems := []models.NavItem{
		{ID: 1, Name: "首页", NameEn: "Home", Slug: "home", NavType: "service", LinkType: "internal", Position: 0, Enabled: true},
		{ID: 2, Name: "沙特签证", NameEn: "Saudi Visa", Slug: "saudi-visa", NavType: "service", LinkType: "internal", ServiceType: "saudi_business_visa", Position: 1, Enabled: true},
		{ID: 3, Name: "全球签证", NameEn: "Global Visa", Slug: "visa", NavType: "service", LinkType: "internal", ServiceType: "other_destinations_visa", Position: 2, Enabled: true},
		{ID: 4, Name: "境外交通住宿", NameEn: "Transport & Accommodation", Slug: "transport", NavType: "service", LinkType: "internal", ServiceType: "transport", Position: 3, Enabled: true},
		{ID: 5, Name: "境外保险", NameEn: "Insurance", Slug: "insurance", NavType: "service", LinkType: "internal", ServiceType: "insurance", Position: 4, Enabled: true},
		{ID: 6, Name: "企业出海", NameEn: "Enterprise Outbound", Slug: "enterprise", NavType: "service", LinkType: "internal", ServiceType: "enterprise_outbound", Position: 5, Enabled: true},
		{ID: 7, Name: "企业考察", NameEn: "Enterprise Inspection", Slug: "inspection", NavType: "service", LinkType: "internal", ServiceType: "enterprise_inspection", Position: 6, Enabled: true},
		{ID: 8, Name: "沙特资讯", NameEn: "Saudi News", Slug: "news", NavType: "blog", LinkType: "internal", Position: 7, Enabled: true},
		{ID: 9, Name: "关于我们", NameEn: "About Us", Slug: "about", NavType: "about", LinkType: "internal", Position: 8, Enabled: true},
	}
	for _, navItem := range defaultNavItems {
		store.navItems[navItem.ID] = &navItem
		store.navSlugMap[navItem.Slug] = navItem.ID
		if navItem.ID > store.navCounter {
			store.navCounter = navItem.ID
		}
	}

	// Create default blog posts
	store.blogs[1] = &models.BlogPost{
		ID:        1,
		Title:     "2024年沙特商务签证新政策解读",
		Slug:      "2024-saudi-business-visa-policy",
		Summary:   "详细解读2024年沙特商务签证的最新政策变化",
		Content:   "沙特阿拉伯政府近期发布了新的商务签证政策...",
		Category:  "政策解读",
		Status:    "published",
		CreatedAt: time.Now().AddDate(0, 0, -5),
		UpdatedAt: time.Now().AddDate(0, 0, -5),
	}
	store.blogSlugMap["2024-saudi-business-visa-policy"] = 1
	store.blogCounter = 1

	// Create default logs
	store.logs = append(store.logs, models.SystemLog{
		ID:          1,
		Action:      "login",
		Description: "Admin user logged in",
		CreatedAt:   time.Now(),
	})
	store.logCounter = 1

	return store
}

func (s *InMemoryStore) CreateUser(user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.userCounter++
	user.ID = s.userCounter
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	s.users[user.ID] = user
	s.usernameIndex[user.Username] = user.ID
	return nil
}

func (s *InMemoryStore) GetUserByUsername(username string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, exists := s.usernameIndex[username]
	if !exists {
		return nil, errors.New("user not found")
	}
	return s.users[id], nil
}

func (s *InMemoryStore) GetUserByID(id uint) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *InMemoryStore) GetAllUsers() []models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var users []models.User
	for _, user := range s.users {
		users = append(users, *user)
	}
	return users
}

func (s *InMemoryStore) DeleteUser(id uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]
	if !exists {
		return errors.New("user not found")
	}
	delete(s.usernameIndex, user.Username)
	delete(s.users, id)
	return nil
}

func (s *InMemoryStore) UpdateUserRole(id uint, role models.Role) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]
	if !exists {
		return errors.New("user not found")
	}
	user.Role = role
	user.UpdatedAt = time.Now()
	return nil
}

func (s *InMemoryStore) CreateContent(content *models.Content) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.contentCounter++
	content.ID = s.contentCounter
	content.CreatedAt = time.Now()
	content.UpdatedAt = time.Now()
	s.contents[content.ID] = content
	return nil
}

func (s *InMemoryStore) GetContentByID(id uint) (*models.Content, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	content, exists := s.contents[id]
	if !exists {
		return nil, errors.New("content not found")
	}
	return content, nil
}

func (s *InMemoryStore) GetContents(userID uint, isAdmin bool) []models.Content {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var contents []models.Content
	for _, content := range s.contents {
		if isAdmin || content.UserID == userID {
			contents = append(contents, *content)
		}
	}
	return contents
}

func (s *InMemoryStore) UpdateContent(content *models.Content) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.contents[content.ID]
	if !exists {
		return errors.New("content not found")
	}
	content.UpdatedAt = time.Now()
	s.contents[content.ID] = content
	return nil
}

func (s *InMemoryStore) DeleteContent(id uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.contents[id]
	if !exists {
		return errors.New("content not found")
	}
	delete(s.contents, id)
	return nil
}

func (s *InMemoryStore) CreateContentBlock(block *models.ContentBlock) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.blockCounter++
	block.ID = s.blockCounter
	block.CreatedAt = time.Now()
	block.UpdatedAt = time.Now()
	s.blocks[block.ID] = block
	return nil
}

func (s *InMemoryStore) CreateBlogPost(blog *models.BlogPost) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.blogSlugMap[blog.Slug]; exists {
		return errors.New("blog post with this slug already exists")
	}

	s.blogCounter++
	blog.ID = s.blogCounter
	blog.CreatedAt = time.Now()
	blog.UpdatedAt = time.Now()
	s.blogs[blog.ID] = blog
	s.blogSlugMap[blog.Slug] = blog.ID
	return nil
}

func (s *InMemoryStore) GetBlogPostBySlug(slug string) (*models.BlogPost, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, exists := s.blogSlugMap[slug]
	if !exists {
		return nil, errors.New("blog post not found")
	}
	return s.blogs[id], nil
}

func (s *InMemoryStore) GetBlogPostByID(id uint) (*models.BlogPost, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	blog, exists := s.blogs[id]
	if !exists {
		return nil, errors.New("blog post not found")
	}
	return blog, nil
}

func (s *InMemoryStore) GetAllBlogPosts() []models.BlogPost {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var blogs []models.BlogPost
	for _, blog := range s.blogs {
		blogs = append(blogs, *blog)
	}
	return blogs
}

func (s *InMemoryStore) BlogSlugExists(slug string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.blogSlugMap[slug]
	return exists
}

func (s *InMemoryStore) UpdateBlogPost(blog *models.BlogPost) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.blogs[blog.ID]
	if !exists {
		return errors.New("blog post not found")
	}

	if existing.Slug != blog.Slug {
		delete(s.blogSlugMap, existing.Slug)
		s.blogSlugMap[blog.Slug] = blog.ID
	}
	blog.UpdatedAt = time.Now()
	s.blogs[blog.ID] = blog
	return nil
}

func (s *InMemoryStore) UpdateBlogPostBySlug(slug string, blog *models.BlogPost) (*models.BlogPost, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, exists := s.blogSlugMap[slug]
	if !exists {
		return nil, errors.New("blog post not found")
	}

	existing := s.blogs[id]
	if existing.Slug != blog.Slug {
		if _, slugExists := s.blogSlugMap[blog.Slug]; slugExists && blog.Slug != slug {
			return nil, errors.New("blog post with this slug already exists")
		}
		delete(s.blogSlugMap, existing.Slug)
		s.blogSlugMap[blog.Slug] = id
	}
	blog.ID = id
	blog.CreatedAt = existing.CreatedAt
	blog.UpdatedAt = time.Now()
	s.blogs[id] = blog
	return blog, nil
}

func (s *InMemoryStore) DeleteBlogPost(id uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	blog, exists := s.blogs[id]
	if !exists {
		return errors.New("blog post not found")
	}
	delete(s.blogSlugMap, blog.Slug)
	delete(s.blogs, id)
	return nil
}

func (s *InMemoryStore) DeleteBlogPostBySlug(slug string) (*models.BlogPost, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, exists := s.blogSlugMap[slug]
	if !exists {
		return nil, errors.New("blog post not found")
	}
	blog := s.blogs[id]
	delete(s.blogSlugMap, slug)
	delete(s.blogs, id)
	return blog, nil
}

func (s *InMemoryStore) GetWebsiteModule(name string) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	module, exists := s.websiteModules[name]
	if !exists {
		return nil, errors.New("website module not found")
	}
	return module, nil
}

func (s *InMemoryStore) UpdateWebsiteModule(name string, data map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.websiteModules[name] = data
	return nil
}

func (s *InMemoryStore) CreateServicePage(page *models.ServicePage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.servicePages[page.Slug]; exists {
		return errors.New("service page with this slug already exists")
	}

	s.serviceCounter++
	page.ID = s.serviceCounter
	page.CreatedAt = time.Now()
	page.UpdatedAt = time.Now()
	s.servicePages[page.Slug] = page
	return nil
}

func (s *InMemoryStore) ServicePageSlugExists(slug string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.servicePages[slug]
	return exists
}

func (s *InMemoryStore) GetServicePageBySlug(slug string) (*models.ServicePage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	page, exists := s.servicePages[slug]
	if !exists {
		return nil, errors.New("service page not found")
	}
	return page, nil
}

func (s *InMemoryStore) GetAllServicePages() []models.ServicePage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var pages []models.ServicePage
	for _, page := range s.servicePages {
		pages = append(pages, *page)
	}
	return pages
}

func (s *InMemoryStore) UpdateServicePage(page *models.ServicePage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.servicePages[page.Slug]
	if !exists {
		// Try to find by ID in case slug changed
		var oldSlug string
		found := false
		for slug, p := range s.servicePages {
			if p.ID == page.ID {
				existing = p
				oldSlug = slug
				found = true
				break
			}
		}
		if !found {
			return errors.New("service page not found")
		}
		if oldSlug != page.Slug {
			if _, slugExists := s.servicePages[page.Slug]; slugExists {
				return errors.New("service page with this slug already exists")
			}
			delete(s.servicePages, oldSlug)
		}
		page.UpdatedAt = time.Now()
		s.servicePages[page.Slug] = page
		return nil
	}

	page.ID = existing.ID
	page.UpdatedAt = time.Now()
	s.servicePages[page.Slug] = page
	return nil
}

func (s *InMemoryStore) DeleteServicePage(slug string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.servicePages[slug]
	if !exists {
		return errors.New("service page not found")
	}
	delete(s.servicePages, slug)
	return nil
}

func (s *InMemoryStore) CreateLog(log *models.SystemLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logCounter++
	log.ID = s.logCounter
	log.CreatedAt = time.Now()
	s.logs = append(s.logs, *log)
	return nil
}

func (s *InMemoryStore) CreateLogWithDetails(action, description string) error {
	return s.CreateLog(&models.SystemLog{
		Action:      action,
		Description: description,
	})
}

func (s *InMemoryStore) GetAllLogs() []models.SystemLog {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.logs
}

func (s *InMemoryStore) UpdateUser(user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.users[user.ID]
	if !exists {
		return errors.New("user not found")
	}

	if existing.Username != user.Username {
		delete(s.usernameIndex, existing.Username)
		s.usernameIndex[user.Username] = user.ID
	}

	user.UpdatedAt = time.Now()
	user.CreatedAt = existing.CreatedAt
	s.users[user.ID] = user
	return nil
}

// Navigation Items CRUD methods
func (s *InMemoryStore) CreateNavItem(item *models.NavItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.navSlugMap[item.Slug]; exists {
		return errors.New("navigation item with this slug already exists")
	}

	s.navCounter++
	item.ID = s.navCounter
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	s.navItems[item.ID] = item
	s.navSlugMap[item.Slug] = item.ID
	return nil
}

func (s *InMemoryStore) GetNavItemBySlug(slug string) (*models.NavItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, exists := s.navSlugMap[slug]
	if !exists {
		return nil, errors.New("navigation item not found")
	}
	return s.navItems[id], nil
}

func (s *InMemoryStore) GetNavItemByID(id uint) (*models.NavItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.navItems[id]
	if !exists {
		return nil, errors.New("navigation item not found")
	}
	return item, nil
}

func (s *InMemoryStore) GetAllNavItems() []models.NavItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var items []models.NavItem
	for _, item := range s.navItems {
		items = append(items, *item)
	}
	return items
}

func (s *InMemoryStore) GetEnabledNavItems() []models.NavItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var items []models.NavItem
	for _, item := range s.navItems {
		if item.Enabled {
			items = append(items, *item)
		}
	}
	return items
}

func (s *InMemoryStore) NavSlugExists(slug string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.navSlugMap[slug]
	return exists
}

func (s *InMemoryStore) UpdateNavItem(item *models.NavItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.navItems[item.ID]
	if !exists {
		return errors.New("navigation item not found")
	}

	if existing.Slug != item.Slug {
		delete(s.navSlugMap, existing.Slug)
		s.navSlugMap[item.Slug] = item.ID
	}
	item.UpdatedAt = time.Now()
	s.navItems[item.ID] = item
	return nil
}

func (s *InMemoryStore) DeleteNavItem(id uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.navItems[id]
	if !exists {
		return errors.New("navigation item not found")
	}
	delete(s.navSlugMap, item.Slug)
	delete(s.navItems, id)
	return nil
}
