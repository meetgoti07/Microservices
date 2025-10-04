-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Categories table
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    image_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL, -- Reference to auth service user
    updated_by UUID
);

-- Menu items table
CREATE TABLE menu_items (
    id SERIAL PRIMARY KEY,
    category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL CHECK (price >= 0),
    image_url VARCHAR(500),
    is_vegetarian BOOLEAN DEFAULT FALSE,
    is_vegan BOOLEAN DEFAULT FALSE,
    contains_allergens VARCHAR(500), -- comma-separated: nuts, dairy, gluten, etc.
    calories INTEGER,
    preparation_time_minutes INTEGER DEFAULT 10,
    is_available BOOLEAN DEFAULT TRUE,
    is_active BOOLEAN DEFAULT TRUE,
    stock_quantity INTEGER DEFAULT NULL, -- NULL means unlimited
    popularity_score INTEGER DEFAULT 0, -- for recommendations
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_by UUID
);

-- Item availability schedule (optional feature)
CREATE TABLE item_availability (
    id SERIAL PRIMARY KEY,
    item_id INTEGER NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL CHECK (day_of_week BETWEEN 0 AND 6), -- 0=Sunday
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Price history (for tracking price changes)
CREATE TABLE price_history (
    id SERIAL PRIMARY KEY,
    item_id INTEGER NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    old_price DECIMAL(10, 2) NOT NULL,
    new_price DECIMAL(10, 2) NOT NULL,
    changed_by UUID NOT NULL,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    reason TEXT
);

-- Menu item reviews/ratings (denormalized from order service)
CREATE TABLE item_ratings (
    id SERIAL PRIMARY KEY,
    item_id INTEGER NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    order_id VARCHAR(50) NOT NULL, -- From order service
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    review TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_categories_active ON categories(is_active);
CREATE INDEX idx_categories_order ON categories(display_order);
CREATE INDEX idx_menu_items_category ON menu_items(category_id);
CREATE INDEX idx_menu_items_available ON menu_items(is_available);
CREATE INDEX idx_menu_items_active ON menu_items(is_active);
CREATE INDEX idx_menu_items_price ON menu_items(price);
CREATE INDEX idx_menu_items_vegetarian ON menu_items(is_vegetarian);
CREATE INDEX idx_item_availability_item ON item_availability(item_id);
CREATE INDEX idx_price_history_item ON price_history(item_id);
CREATE INDEX idx_item_ratings_item ON item_ratings(item_id);
CREATE INDEX idx_item_ratings_user ON item_ratings(user_id);

-- Triggers
CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_menu_items_updated_at BEFORE UPDATE ON menu_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to update popularity score
CREATE OR REPLACE FUNCTION update_popularity_score()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE menu_items 
    SET popularity_score = popularity_score + 1
    WHERE id = NEW.item_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Seed data
INSERT INTO categories (name, description, display_order, created_by) VALUES
('Breakfast', 'Morning delights', 1, '00000000-0000-0000-0000-000000000000'),
('Main Course', 'Hearty meals', 2, '00000000-0000-0000-0000-000000000000'),
('Snacks', 'Quick bites', 3, '00000000-0000-0000-0000-000000000000'),
('Beverages', 'Refreshing drinks', 4, '00000000-0000-0000-0000-000000000000'),
('Desserts', 'Sweet treats', 5, '00000000-0000-0000-0000-000000000000');

INSERT INTO menu_items (category_id, name, description, price, is_vegetarian, preparation_time_minutes, created_by) VALUES
(1, 'Idli Sambar', 'Steamed rice cakes with sambar', 40.00, TRUE, 10, '00000000-0000-0000-0000-000000000000'),
(1, 'Masala Dosa', 'Crispy dosa with potato filling', 50.00, TRUE, 15, '00000000-0000-0000-0000-000000000000'),
(2, 'Veg Thali', 'Complete vegetarian meal', 120.00, TRUE, 20, '00000000-0000-0000-0000-000000000000'),
(2, 'Paneer Butter Masala', 'Rich paneer curry with naan', 150.00, TRUE, 25, '00000000-0000-0000-0000-000000000000'),
(3, 'Samosa', 'Crispy fried pastry', 20.00, TRUE, 5, '00000000-0000-0000-0000-000000000000'),
(4, 'Masala Chai', 'Indian spiced tea', 15.00, TRUE, 5, '00000000-0000-0000-0000-000000000000'),
(4, 'Cold Coffee', 'Chilled coffee drink', 40.00, TRUE, 5, '00000000-0000-0000-0000-000000000000'),
(5, 'Gulab Jamun', 'Sweet milk solid balls', 30.00, TRUE, 5, '00000000-0000-0000-0000-000000000000');