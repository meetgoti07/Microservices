-- ============================================
-- Queue Service Database Schema
-- ============================================
-- This service manages queue entries, tokens, and queue configuration
-- References userId from auth service and orderId from order service (no FK constraints)

-- Enable UUID extension if using PostgreSQL
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- Queue Entries Table
-- ============================================
CREATE TABLE IF NOT EXISTS queue_entries (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    order_id VARCHAR(36) UNIQUE NOT NULL, -- Reference to order service (no FK)
    user_id VARCHAR(36) NOT NULL, -- Reference to auth service (no FK)
    
    -- User info cached from auth service
    user_name VARCHAR(200),
    user_phone VARCHAR(20),
    
    -- Token information
    token_number VARCHAR(20) UNIQUE NOT NULL, -- e.g., "A001", "B123"
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
    estimated_wait_time INT DEFAULT 0, -- minutes
    estimated_ready_time TIMESTAMP,
    
    -- Actual timing
    actual_start_time TIMESTAMP,
    actual_ready_time TIMESTAMP,
    actual_completion_time TIMESTAMP,
    
    -- Assignment
    assigned_counter VARCHAR(50),
    assigned_staff VARCHAR(36), -- user_id of staff
    assigned_staff_name VARCHAR(100),
    
    -- Preparation
    average_item_preparation_time INT, -- minutes
    
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
-- Queue Notifications Sent Table
-- ============================================
CREATE TABLE IF NOT EXISTS queue_notifications_sent (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    queue_entry_id VARCHAR(36) NOT NULL,
    notification_type ENUM(
        'ORDER_CONFIRMED', 'POSITION_UPDATE', 
        'ALMOST_READY', 'READY', 'REMINDER'
    ) NOT NULL,
    channel ENUM('IN_APP',  'EMAIL') NOT NULL,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_queue_entry_id (queue_entry_id),
    INDEX idx_notification_type (notification_type),
    INDEX idx_sent_at (sent_at),
    
    FOREIGN KEY (queue_entry_id) REFERENCES queue_entries(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Queue Position History Table
-- ============================================
CREATE TABLE IF NOT EXISTS queue_position_history (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    queue_entry_id VARCHAR(36) NOT NULL,
    old_position INT NOT NULL,
    new_position INT NOT NULL,
    old_status ENUM(
        'WAITING', 'IN_PROGRESS', 'READY', 
        'COMPLETED', 'CANCELLED', 'NO_SHOW', 'EXPIRED'
    ) NOT NULL,
    new_status ENUM(
        'WAITING', 'IN_PROGRESS', 'READY', 
        'COMPLETED', 'CANCELLED', 'NO_SHOW', 'EXPIRED'
    ) NOT NULL,
    estimated_wait_time INT,
    estimated_ready_time TIMESTAMP,
    reason TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_queue_entry_id (queue_entry_id),
    INDEX idx_timestamp (timestamp),
    
    FOREIGN KEY (queue_entry_id) REFERENCES queue_entries(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Queue Configuration Table
-- ============================================
CREATE TABLE IF NOT EXISTS queue_configuration (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    
    -- Capacity settings
    max_concurrent_orders INT DEFAULT 10 CHECK (max_concurrent_orders > 0),
    avg_preparation_time_per_item INT DEFAULT 5 CHECK (avg_preparation_time_per_item > 0), -- minutes
    buffer_time INT DEFAULT 2 CHECK (buffer_time >= 0), -- minutes
    
    -- Express queue settings
    express_queue_enabled BOOLEAN DEFAULT FALSE,
    express_queue_max_items INT DEFAULT 3 CHECK (express_queue_max_items > 0),
    
    -- Alert settings
    max_wait_time_alert INT DEFAULT 30, -- minutes
    token_expiry_time INT DEFAULT 60, -- minutes
    
    -- Notification settings
    auto_notification_enabled BOOLEAN DEFAULT TRUE,
    notification_position_threshold INT DEFAULT 5,
    notification_almost_ready_threshold INT DEFAULT 2, -- minutes
    
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    updated_by VARCHAR(36) -- user_id
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Queue Working Hours Table
-- ============================================
CREATE TABLE IF NOT EXISTS queue_working_hours (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    configuration_id VARCHAR(36) NOT NULL,
    day ENUM(
        'MONDAY', 'TUESDAY', 'WEDNESDAY', 
        'THURSDAY', 'FRIDAY', 'SATURDAY', 'SUNDAY'
    ) NOT NULL,
    open_time VARCHAR(5) NOT NULL, -- HH:MM format
    close_time VARCHAR(5) NOT NULL, -- HH:MM format
    is_open BOOLEAN DEFAULT TRUE,
    
    INDEX idx_configuration_id (configuration_id),
    INDEX idx_day (day),
    UNIQUE KEY unique_config_day (configuration_id, day),
    
    FOREIGN KEY (configuration_id) REFERENCES queue_configuration(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Queue Priority Multipliers Table
-- ============================================
CREATE TABLE IF NOT EXISTS queue_priority_multipliers (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    configuration_id VARCHAR(36) NOT NULL,
    priority ENUM('LOW', 'NORMAL', 'HIGH', 'URGENT', 'VIP') NOT NULL,
    multiplier DECIMAL(3, 2) DEFAULT 1.00 CHECK (multiplier BETWEEN 0.1 AND 10.0),
    
    INDEX idx_configuration_id (configuration_id),
    UNIQUE KEY unique_config_priority (configuration_id, priority),
    
    FOREIGN KEY (configuration_id) REFERENCES queue_configuration(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Queue Display Announcements Table
-- ============================================
CREATE TABLE IF NOT EXISTS queue_display_announcements (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    message TEXT NOT NULL,
    type ENUM('INFO', 'WARNING', 'URGENT') DEFAULT 'INFO',
    priority INT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    display_until TIMESTAMP,
    created_by VARCHAR(36), -- user_id
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_is_active (is_active),
    INDEX idx_display_until (display_until),
    INDEX idx_priority (priority)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Staff Queue Actions Log Table
-- ============================================
CREATE TABLE IF NOT EXISTS staff_queue_actions_log (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    queue_entry_id VARCHAR(36) NOT NULL,
    staff_id VARCHAR(36) NOT NULL, -- user_id
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
    INDEX idx_timestamp (timestamp),
    
    FOREIGN KEY (queue_entry_id) REFERENCES queue_entries(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Queue Statistics Table (Daily aggregates)
-- ============================================
CREATE TABLE IF NOT EXISTS queue_statistics (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
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
    avg_wait_time INT DEFAULT 0, -- minutes
    avg_preparation_time INT DEFAULT 0, -- minutes
    longest_wait_time INT DEFAULT 0, -- minutes
    shortest_wait_time INT DEFAULT 0, -- minutes
    
    -- Capacity metrics
    current_load DECIMAL(5, 2) DEFAULT 0.00, -- percentage
    peak_load DECIMAL(5, 2) DEFAULT 0.00, -- percentage
    peak_load_time VARCHAR(5), -- HH:MM
    
    -- Customer satisfaction
    on_time_completion_rate DECIMAL(5, 2) DEFAULT 0.00, -- percentage
    no_show_rate DECIMAL(5, 2) DEFAULT 0.00, -- percentage
    
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_date (date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Queue Hourly Statistics Table
-- ============================================
CREATE TABLE IF NOT EXISTS queue_hourly_statistics (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    date DATE NOT NULL,
    hour INT NOT NULL CHECK (hour BETWEEN 0 AND 23),
    
    -- Hourly metrics
    order_count INT DEFAULT 0,
    avg_wait_time INT DEFAULT 0, -- minutes
    avg_preparation_time INT DEFAULT 0, -- minutes
    completed_count INT DEFAULT 0,
    cancelled_count INT DEFAULT 0,
    peak_position INT DEFAULT 0,
    
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_date (date),
    INDEX idx_hour (hour),
    UNIQUE KEY unique_date_hour (date, hour)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Queue Token Counter Table (For generating sequential tokens)
-- ============================================
CREATE TABLE IF NOT EXISTS queue_token_counter (
    id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
    date DATE UNIQUE NOT NULL,
    current_number INT DEFAULT 0,
    prefix VARCHAR(1) DEFAULT 'A',
    last_reset_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_date (date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================
-- Insert Default Configuration
-- ============================================
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
    UUID(),
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

-- Get the configuration ID for working hours and priority multipliers
SET @config_id = (SELECT id FROM queue_configuration LIMIT 1);

-- Insert default working hours (open 8:00 AM to 10:00 PM every day)
INSERT INTO queue_working_hours (id, configuration_id, day, open_time, close_time, is_open) VALUES
    (UUID(), @config_id, 'MONDAY', '08:00', '22:00', TRUE),
    (UUID(), @config_id, 'TUESDAY', '08:00', '22:00', TRUE),
    (UUID(), @config_id, 'WEDNESDAY', '08:00', '22:00', TRUE),
    (UUID(), @config_id, 'THURSDAY', '08:00', '22:00', TRUE),
    (UUID(), @config_id, 'FRIDAY', '08:00', '22:00', TRUE),
    (UUID(), @config_id, 'SATURDAY', '08:00', '22:00', TRUE),
    (UUID(), @config_id, 'SUNDAY', '08:00', '22:00', TRUE);

-- Insert default priority multipliers
INSERT INTO queue_priority_multipliers (id, configuration_id, priority, multiplier) VALUES
    (UUID(), @config_id, 'LOW', 1.50),
    (UUID(), @config_id, 'NORMAL', 1.00),
    (UUID(), @config_id, 'HIGH', 0.70),
    (UUID(), @config_id, 'URGENT', 0.50),
    (UUID(), @config_id, 'VIP', 0.30);

-- Insert initial token counter for today
INSERT INTO queue_token_counter (id, date, current_number, prefix) 
VALUES (UUID(), CURDATE(), 0, 'A');
