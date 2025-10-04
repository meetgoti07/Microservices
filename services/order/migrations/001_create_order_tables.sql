-- ============================================
-- Order Service Database Schema
-- ============================================
-- This service manages orders, carts, and payments
-- References userId from auth service and menuItemId from menu service (no FK constraints)

-- Enable UUID extension if using PostgreSQL
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- Shopping Carts Table
-- ============================================
CREATE TABLE IF NOT EXISTS shopping_carts (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id VARCHAR(36) UNIQUE NOT NULL, -- Reference to auth service (no FK)
    
    -- Totals (calculated)
    subtotal DECIMAL(10, 2) DEFAULT 0.00,
    tax DECIMAL(10, 2) DEFAULT 0.00,
    tax_percentage DECIMAL(5, 2) DEFAULT 0.00,
    total DECIMAL(10, 2) DEFAULT 0.00,
    item_count INT DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_user_id (user_id),
    INDEX idx_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Cart Items Table
-- ============================================
CREATE TABLE IF NOT EXISTS cart_items (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    cart_id VARCHAR(36) NOT NULL,
    menu_item_id VARCHAR(36) NOT NULL, -- Reference to menu service (no FK)
    
    -- Menu item info cached from menu service
    menu_item_name VARCHAR(200),
    menu_item_image VARCHAR(500),
    
    quantity INT NOT NULL CHECK (quantity BETWEEN 1 AND 50),
    unit_price DECIMAL(10, 2) NOT NULL,
    total_price DECIMAL(10, 2) NOT NULL,
    special_instructions TEXT,
    
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_cart_id (cart_id),
    INDEX idx_menu_item_id (menu_item_id),
    INDEX idx_added_at (added_at),
    
    FOREIGN KEY (cart_id) REFERENCES shopping_carts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Cart Item Customizations Table
-- ============================================
CREATE TABLE IF NOT EXISTS cart_item_customizations (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    cart_item_id VARCHAR(36) NOT NULL,
    customization_id VARCHAR(100) NOT NULL,
    customization_name VARCHAR(100) NOT NULL,
    selected_label VARCHAR(100) NOT NULL,
    selected_value VARCHAR(100) NOT NULL,
    price_modifier DECIMAL(10, 2) DEFAULT 0.00,
    
    INDEX idx_cart_item_id (cart_item_id),
    
    FOREIGN KEY (cart_item_id) REFERENCES cart_items(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Orders Table
-- ============================================
CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    order_number VARCHAR(50) UNIQUE NOT NULL,
    user_id VARCHAR(36) NOT NULL, -- Reference to auth service (no FK)
    
    -- User info cached from auth service
    user_name VARCHAR(200),
    user_email VARCHAR(255),
    user_phone VARCHAR(20),
    
    status ENUM(
        'CART', 'PENDING', 'CONFIRMED', 
        'PREPARING', 'READY', 'COMPLETED', 
        'CANCELLED', 'REFUNDED'
    ) DEFAULT 'PENDING',
    
    -- Totals
    subtotal DECIMAL(10, 2) NOT NULL,
    tax DECIMAL(10, 2) DEFAULT 0.00,
    total DECIMAL(10, 2) NOT NULL,

    
    -- Special instructions
    special_instructions TEXT,
    
    -- Time tracking
    estimated_preparation_time INT DEFAULT 0, -- minutes
    actual_preparation_time INT, -- minutes
    estimated_ready_time TIMESTAMP,
    actual_ready_time TIMESTAMP,
    completed_at TIMESTAMP,
    
    -- Cancellation
    cancelled_at TIMESTAMP,
    cancellation_reason TEXT,
    cancelled_by VARCHAR(36), -- user_id
    
    -- Feedback
    rating INT CHECK (rating BETWEEN 1 AND 5),
    feedback TEXT,
    feedback_submitted_at TIMESTAMP,
    
    -- Metadata
    metadata JSON,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_order_number (order_number),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    INDEX idx_estimated_ready_time (estimated_ready_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Order Tokens Table
-- ============================================
CREATE TABLE IF NOT EXISTS order_tokens (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    order_id VARCHAR(36) UNIQUE NOT NULL,
    token VARCHAR(20) UNIQUE NOT NULL, -- e.g., "A001", "B123"
    token_number INT NOT NULL,
    token_prefix VARCHAR(1),
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    
    INDEX idx_order_id (order_id),
    INDEX idx_token (token),
    INDEX idx_generated_at (generated_at),
    
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Order Items Table
-- ============================================
CREATE TABLE IF NOT EXISTS order_items (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    order_id VARCHAR(36) NOT NULL,
    menu_item_id VARCHAR(36) NOT NULL, -- Reference to menu service (no FK)
    
    -- Menu item info cached from menu service
    menu_item_name VARCHAR(200) NOT NULL,
    menu_item_image VARCHAR(500),
    
    quantity INT NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(10, 2) NOT NULL,
    total_price DECIMAL(10, 2) NOT NULL,
    special_instructions TEXT,
    
    INDEX idx_order_id (order_id),
    INDEX idx_menu_item_id (menu_item_id),
    
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Order Item Customizations Table
-- ============================================
CREATE TABLE IF NOT EXISTS order_item_customizations (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    order_item_id VARCHAR(36) NOT NULL,
    customization_id VARCHAR(100) NOT NULL,
    customization_name VARCHAR(100) NOT NULL,
    selected_label VARCHAR(100) NOT NULL,
    selected_value VARCHAR(100) NOT NULL,
    price_modifier DECIMAL(10, 2) DEFAULT 0.00,
    
    INDEX idx_order_item_id (order_item_id),
    
    FOREIGN KEY (order_item_id) REFERENCES order_items(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Order Timeline Table
-- ============================================
CREATE TABLE IF NOT EXISTS order_timeline (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    order_id VARCHAR(36) NOT NULL,
    status ENUM(
        'CART', 'PENDING', 'CONFIRMED', 
        'PREPARING', 'READY', 'COMPLETED', 
        'CANCELLED', 'REFUNDED'
    ) NOT NULL,
    message TEXT,
    updated_by VARCHAR(36), -- user_id
    updated_by_name VARCHAR(100),
    metadata JSON,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_order_id (order_id),
    INDEX idx_status (status),
    INDEX idx_timestamp (timestamp),
    
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Payments Table
-- ============================================
CREATE TABLE IF NOT EXISTS payments (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    order_id VARCHAR(36) UNIQUE NOT NULL,
    
    method ENUM(
        'CASH', 'UPI', 'CARD', 
        'WALLET', 'CAMPUS_CARD', 'MEAL_COUPON'
    ) NOT NULL,
    
    status ENUM(
        'PENDING', 'PROCESSING', 'COMPLETED', 
        'FAILED', 'REFUNDED', 'CANCELLED'
    ) DEFAULT 'PENDING',
    
    amount DECIMAL(10, 2) NOT NULL,
    
    -- Payment gateway details
    transaction_id VARCHAR(200),
    transaction_reference VARCHAR(200),
    payment_gateway VARCHAR(100),
    
    -- Payment method specific fields
    upi_id VARCHAR(100),
    card_last_4_digits VARCHAR(4),
    card_type ENUM('VISA', 'MASTERCARD', 'RUPAY', 'AMEX'),
    
    -- Timing
    initiated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    
    -- Failure handling
    failure_reason TEXT,
    retry_count INT DEFAULT 0,
    
    -- Metadata
    metadata JSON,
    
    INDEX idx_order_id (order_id),
    INDEX idx_status (status),
    INDEX idx_method (method),
    INDEX idx_transaction_id (transaction_id),
    INDEX idx_initiated_at (initiated_at),
    
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Order Item Feedback Table
-- ============================================
CREATE TABLE IF NOT EXISTS order_item_feedback (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    order_id VARCHAR(36) NOT NULL,
    order_item_id VARCHAR(36) NOT NULL,
    rating INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_order_id (order_id),
    INDEX idx_order_item_id (order_item_id),
    INDEX idx_rating (rating),
    
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY (order_item_id) REFERENCES order_items(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


-- ============================================
-- Order Statistics Table (Daily aggregates)
-- ============================================
CREATE TABLE IF NOT EXISTS order_statistics (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    date DATE UNIQUE NOT NULL,
    total_orders INT DEFAULT 0,
    pending_orders INT DEFAULT 0,
    confirmed_orders INT DEFAULT 0,
    preparing_orders INT DEFAULT 0,
    ready_orders INT DEFAULT 0,
    completed_orders INT DEFAULT 0,
    cancelled_orders INT DEFAULT 0,
    refunded_orders INT DEFAULT 0,
    total_revenue DECIMAL(10, 2) DEFAULT 0.00,
    total_tax DECIMAL(10, 2) DEFAULT 0.00,
    avg_order_value DECIMAL(10, 2) DEFAULT 0.00,
    avg_preparation_time INT DEFAULT 0, -- minutes
    
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_date (date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
