-- ============================================
-- Menu Service Database Schema
-- ============================================
-- This service manages menu items, categories, and reviews
-- References userId from auth service (no FK constraint)

-- Enable UUID extension if using PostgreSQL
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- Menu Items Table
-- ============================================
CREATE TABLE IF NOT EXISTS menu_items (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    short_description VARCHAR(200),
    category ENUM(
        'BREAKFAST', 'LUNCH', 'DINNER', 
        'BEVERAGES', 'SNACKS', 'DESSERTS',
        'COMBO_MEALS', 'HEALTHY_OPTIONS', 'SPECIAL_OF_THE_DAY'
    ) NOT NULL,
    sub_category VARCHAR(100),
    
    -- Pricing
    price DECIMAL(10, 2) NOT NULL,
    discounted_price DECIMAL(10, 2),
    currency VARCHAR(3) DEFAULT 'INR',
    
    -- Availability
    availability_status ENUM(
        'AVAILABLE', 'OUT_OF_STOCK', 
        'COMING_SOON', 'TEMPORARILY_UNAVAILABLE', 'SEASONAL'
    ) DEFAULT 'AVAILABLE',
    stock_quantity INT DEFAULT 0,
    low_stock_threshold INT DEFAULT 10,
    
    -- Preparation
    preparation_time INT NOT NULL DEFAULT 10, -- in minutes
    serving_size VARCHAR(100),
    
    -- Properties
    spice_level ENUM('NONE', 'MILD', 'MEDIUM', 'HOT', 'EXTRA_HOT'),
    is_popular BOOLEAN DEFAULT FALSE,
    is_featured BOOLEAN DEFAULT FALSE,
    is_new_item BOOLEAN DEFAULT FALSE,
    
    -- Ratings
    rating DECIMAL(3, 2) DEFAULT 0.00,
    review_count INT DEFAULT 0,
    
    -- Metadata
    tags JSON, -- Array of tags
    available_days JSON, -- Array of days
    available_time_slots JSON, -- Array of time slot objects
    
    -- User references (no FK - managed by auth service)
    created_by VARCHAR(36),
    updated_by VARCHAR(36),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_name (name),
    INDEX idx_category (category),
    INDEX idx_availability_status (availability_status),
    INDEX idx_price (price),
    INDEX idx_rating (rating),
    INDEX idx_is_popular (is_popular),
    INDEX idx_is_featured (is_featured),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Dietary Restrictions Table
-- ============================================
CREATE TABLE IF NOT EXISTS menu_item_dietary_restrictions (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    menu_item_id VARCHAR(36) NOT NULL,
    restriction ENUM(
        'VEGETARIAN', 'VEGAN', 'GLUTEN_FREE', 
        'NUT_FREE', 'DAIRY_FREE', 'HALAL', 
        'JAIN', 'KETO', 'LOW_CARB', 'HIGH_PROTEIN'
    ) NOT NULL,
    
    INDEX idx_menu_item_id (menu_item_id),
    INDEX idx_restriction (restriction),
    UNIQUE KEY unique_item_restriction (menu_item_id, restriction),
    
    FOREIGN KEY (menu_item_id) REFERENCES menu_items(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Allergens Table
-- ============================================
CREATE TABLE IF NOT EXISTS menu_item_allergens (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    menu_item_id VARCHAR(36) NOT NULL,
    allergen VARCHAR(100) NOT NULL,
    
    INDEX idx_menu_item_id (menu_item_id),
    INDEX idx_allergen (allergen),
    UNIQUE KEY unique_item_allergen (menu_item_id, allergen),
    
    FOREIGN KEY (menu_item_id) REFERENCES menu_items(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Ingredients Table
-- ============================================
CREATE TABLE IF NOT EXISTS menu_item_ingredients (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    menu_item_id VARCHAR(36) NOT NULL,
    ingredient VARCHAR(200) NOT NULL,
    display_order INT DEFAULT 0,
    
    INDEX idx_menu_item_id (menu_item_id),
    
    FOREIGN KEY (menu_item_id) REFERENCES menu_items(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Nutritional Information Table
-- ============================================
CREATE TABLE IF NOT EXISTS menu_item_nutrition (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    menu_item_id VARCHAR(36) UNIQUE NOT NULL,
    calories DECIMAL(8, 2),
    protein DECIMAL(8, 2), -- grams
    carbohydrates DECIMAL(8, 2), -- grams
    fat DECIMAL(8, 2), -- grams
    fiber DECIMAL(8, 2), -- grams
    sodium DECIMAL(8, 2), -- mg
    sugar DECIMAL(8, 2), -- grams
    
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_menu_item_id (menu_item_id),
    
    FOREIGN KEY (menu_item_id) REFERENCES menu_items(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Customization Options Table
-- ============================================
CREATE TABLE IF NOT EXISTS menu_item_customizations (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    menu_item_id VARCHAR(36) NOT NULL,
    customization_id VARCHAR(100) NOT NULL, -- e.g., "spice_level", "size"
    name VARCHAR(100) NOT NULL,
    is_required BOOLEAN DEFAULT FALSE,
    max_selections INT DEFAULT 1,
    display_order INT DEFAULT 0,
    
    INDEX idx_menu_item_id (menu_item_id),
    UNIQUE KEY unique_item_customization (menu_item_id, customization_id),
    
    FOREIGN KEY (menu_item_id) REFERENCES menu_items(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Customization Options Values Table
-- ============================================
CREATE TABLE IF NOT EXISTS menu_item_customization_options (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    customization_id VARCHAR(36) NOT NULL,
    label VARCHAR(100) NOT NULL,
    value VARCHAR(100) NOT NULL,
    price_modifier DECIMAL(10, 2) DEFAULT 0.00,
    display_order INT DEFAULT 0,
    
    INDEX idx_customization_id (customization_id),
    
    FOREIGN KEY (customization_id) REFERENCES menu_item_customizations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Menu Item Reviews Table
-- ============================================
CREATE TABLE IF NOT EXISTS menu_item_reviews (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    menu_item_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL, -- Reference to auth service (no FK)
    order_id VARCHAR(36), -- Reference to order service (no FK)
    rating INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT,
    
    -- User info cached from auth service
    user_name VARCHAR(100),
    user_avatar VARCHAR(500),
    
    -- Moderation
    is_verified_purchase BOOLEAN DEFAULT FALSE,
    is_approved BOOLEAN DEFAULT TRUE,
    is_flagged BOOLEAN DEFAULT FALSE,
    moderated_by VARCHAR(36),
    moderation_notes TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_menu_item_id (menu_item_id),
    INDEX idx_user_id (user_id),
    INDEX idx_order_id (order_id),
    INDEX idx_rating (rating),
    INDEX idx_is_approved (is_approved),
    INDEX idx_created_at (created_at),
    
    FOREIGN KEY (menu_item_id) REFERENCES menu_items(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


-- ============================================
-- Menu Analytics Table
-- ============================================
CREATE TABLE IF NOT EXISTS menu_item_analytics (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    menu_item_id VARCHAR(36) NOT NULL,
    date DATE NOT NULL,
    
    -- Order statistics
    total_orders INT DEFAULT 0,
    total_quantity_sold INT DEFAULT 0,
    total_revenue DECIMAL(10, 2) DEFAULT 0.00,
    
    -- Performance metrics
    avg_preparation_time INT DEFAULT 0, -- minutes
    out_of_stock_duration INT DEFAULT 0, -- minutes
    
    -- Reviews
    reviews_count INT DEFAULT 0,
    avg_rating DECIMAL(3, 2) DEFAULT 0.00,
    
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_menu_item_id (menu_item_id),
    INDEX idx_date (date),
    UNIQUE KEY unique_item_date (menu_item_id, date),
    
    FOREIGN KEY (menu_item_id) REFERENCES menu_items(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Menu Categories Configuration Table
-- ============================================
CREATE TABLE IF NOT EXISTS menu_categories_config (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    category VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    description TEXT,
    icon VARCHAR(200),
    display_order INT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    available_time_start VARCHAR(5), -- HH:MM
    available_time_end VARCHAR(5), -- HH:MM
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_category (category),
    INDEX idx_is_active (is_active),
    INDEX idx_display_order (display_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Insert Default Categories
-- ============================================
INSERT INTO menu_categories_config (id, category, display_name, description, display_order, is_active) VALUES
    (UUID(), 'BREAKFAST', 'Breakfast', 'Start your day right with our breakfast menu', 1, TRUE),
    (UUID(), 'LUNCH', 'Lunch', 'Wholesome lunch options', 2, TRUE),
    (UUID(), 'DINNER', 'Dinner', 'Delicious dinner meals', 3, TRUE),
    (UUID(), 'BEVERAGES', 'Beverages', 'Refreshing drinks and beverages', 4, TRUE),
    (UUID(), 'SNACKS', 'Snacks', 'Quick bites and snacks', 5, TRUE),
    (UUID(), 'DESSERTS', 'Desserts', 'Sweet treats and desserts', 6, TRUE),
    (UUID(), 'COMBO_MEALS', 'Combo Meals', 'Value combo meals', 7, TRUE),
    (UUID(), 'HEALTHY_OPTIONS', 'Healthy Options', 'Nutritious and healthy choices', 8, TRUE),
    (UUID(), 'SPECIAL_OF_THE_DAY', 'Today\'s Special', 'Special items of the day', 9, TRUE);
