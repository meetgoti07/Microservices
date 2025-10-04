-- ============================================
-- Queue Service Database Schema (MySQL Compatible)
-- ============================================

-- ============================================
-- Queue Configuration Table (Create First - No FK Dependencies)
-- ============================================
CREATE TABLE IF NOT EXISTS queue_configuration (
    id VARCHAR(36) PRIMARY KEY,
    
    -- Capacity settings
    max_concurrent_orders INT DEFAULT 10 CHECK (max_concurrent_orders > 0),
    avg_preparation_time_per_item INT DEFAULT 5 CHECK (avg_preparation_time_per_item > 0),
    buffer_time INT DEFAULT 2 CHECK (buffer_time >= 0),
    
    -- Express queue settings
    express_queue_enabled BOOLEAN DEFAULT FALSE,
    express_queue_max_items INT DEFAULT 3 CHECK (express_queue_max_items > 0),
    
    -- Alert settings
    max_wait_time_alert INT DEFAULT 30,
    token_expiry_time INT DEFAULT 60,
    
    -- Notification settings
    auto_notification_enabled BOOLEAN DEFAULT TRUE,
    notification_position_threshold INT DEFAULT 5,
    notification_almost_ready_threshold INT DEFAULT 2,
    
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    updated_by VARCHAR(36)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert Default Configuration
INSERT INTO queue_configuration (
    id, 
    max_concurrent_orders, 
    avg_preparation_time_per_item, 
    buffer_time,
    express_queue_enabled,
    express_queue_max_items,
    max_wait_time_alert,
    token_expiry_time,
    auto_notification_enabled,
    notification_position_threshold,
    notification_almost_ready_threshold
) VALUES (
    '00000000-0000-0000-0000-000000000001',
    10,
    5,
    2,
    FALSE,
    3,
    30,
    60,
    TRUE,
    5,
    2
);

-- ============================================
-- Queue Entries Table
-- ============================================
CREATE TABLE IF NOT EXISTS queue_entries (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) UNIQUE NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    
    -- User info
    user_name VARCHAR(200),
    user_phone VARCHAR(20),
    
    -- Token information
    token_number VARCHAR(20) UNIQUE NOT NULL,
    token_type ENUM('REGULAR', 'EXPRESS', 'BULK', 'SPECIAL', 'STAFF') DEFAULT 'REGULAR',
    
    -- Queue status
    status ENUM(
        'WAITING', 'IN_PROGRESS', 'READY', 
        'COMPLETED', 'CANCELLED', 'NO_SHOW', 'EXPIRED'
    ) DEFAULT 'WAITING',
    
    -- Priority
    priority ENUM('LOW', 'NORMAL', 'HIGH', 'URGENT', 'VIP') DEFAULT 'NORMAL',
    
    -- Position and timing
    position INT NOT NULL CHECK (position > 0),
    estimated_wait_time INT DEFAULT 0,
    estimated_ready_time TIMESTAMP NULL,
    
    -- Actual timing
    actual_start_time TIMESTAMP NULL,
    actual_ready_time TIMESTAMP NULL,
    actual_completion_time TIMESTAMP NULL,
    
    -- Assignment
    assigned_counter VARCHAR(50),
    assigned_staff VARCHAR(36),
    assigned_staff_name VARCHAR(100),
    
    -- Preparation
    average_item_preparation_time INT,
    
    -- Special handling
    is_express_queue BOOLEAN DEFAULT FALSE,
    special_handling TEXT,
    notes TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_order_id (order_id),
    INDEX idx_user_id (user_id),
    INDEX idx_token_number (token_number),
    INDEX idx_status (status),
    INDEX idx_priority (priority),
    INDEX idx_position (position),
    INDEX idx_assigned_staff (assigned_staff),
    INDEX idx_assigned_counter (assigned_counter),
    INDEX idx_created_at (created_at),
    INDEX idx_estimated_ready_time (estimated_ready_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Queue Statistics Table (Daily aggregates)
-- ============================================
CREATE TABLE IF NOT EXISTS queue_statistics (
    id VARCHAR(36) PRIMARY KEY,
    date DATE UNIQUE NOT NULL,
    
    -- Queue counts
    total_in_queue INT DEFAULT 0,
    waiting_count INT DEFAULT 0,
    in_progress_count INT DEFAULT 0,
    ready_count INT DEFAULT 0,
    completed_today INT DEFAULT 0,
    cancelled_today INT DEFAULT 0,
    no_show_today INT DEFAULT 0,
    expired_today INT DEFAULT 0,
    
    -- Performance metrics
    avg_wait_time INT DEFAULT 0,
    avg_preparation_time INT DEFAULT 0,
    longest_wait_time INT DEFAULT 0,
    shortest_wait_time INT DEFAULT 0,
    
    -- Capacity metrics
    current_load DECIMAL(5, 2) DEFAULT 0.00,
    peak_load DECIMAL(5, 2) DEFAULT 0.00,
    peak_load_time VARCHAR(5),
    
    -- Customer satisfaction
    on_time_completion_rate DECIMAL(5, 2) DEFAULT 0.00,
    no_show_rate DECIMAL(5, 2) DEFAULT 0.00,
    
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_date (date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Queue Token Counter Table
-- ============================================
CREATE TABLE IF NOT EXISTS queue_token_counter (
    id VARCHAR(36) PRIMARY KEY,
    date DATE UNIQUE NOT NULL,
    current_number INT DEFAULT 0,
    prefix VARCHAR(1) DEFAULT 'A',
    last_reset_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_date (date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert initial token counter for today
INSERT INTO queue_token_counter (id, date, current_number, prefix) 
VALUES ('10000000-0000-0000-0000-000000000001', CURDATE(), 0, 'A');

-- ============================================
-- Staff Queue Actions Log Table
-- ============================================
CREATE TABLE IF NOT EXISTS staff_queue_actions_log (
    id VARCHAR(36) PRIMARY KEY,
    queue_entry_id VARCHAR(36) NOT NULL,
    staff_id VARCHAR(36) NOT NULL,
    staff_name VARCHAR(100),
    action ENUM(
        'START_PREPARATION', 'MARK_READY', 'MARK_COMPLETED',
        'CANCEL', 'REASSIGN', 'ADJUST_PRIORITY', 'ADD_NOTE'
    ) NOT NULL,
    old_status ENUM(
        'WAITING', 'IN_PROGRESS', 'READY', 
        'COMPLETED', 'CANCELLED', 'NO_SHOW', 'EXPIRED'
    ),
    new_status ENUM(
        'WAITING', 'IN_PROGRESS', 'READY', 
        'COMPLETED', 'CANCELLED', 'NO_SHOW', 'EXPIRED'
    ),
    old_priority ENUM('LOW', 'NORMAL', 'HIGH', 'URGENT', 'VIP'),
    new_priority ENUM('LOW', 'NORMAL', 'HIGH', 'URGENT', 'VIP'),
    assigned_counter VARCHAR(50),
    assigned_staff VARCHAR(36),
    note TEXT,
    reason TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_queue_entry_id (queue_entry_id),
    INDEX idx_staff_id (staff_id),
    INDEX idx_action (action),
    INDEX idx_timestamp (timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
