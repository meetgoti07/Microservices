package com.example.map.model;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;

@Entity
@Table(name = "menu_items")
@Data
@NoArgsConstructor
@AllArgsConstructor
public class MenuItem {

    @Id
    @Column(length = 36)
    private String id;

    @Column(nullable = false, length = 200)
    private String name;

    @Column(columnDefinition = "TEXT")
    private String description;

    @Column(name = "short_description", length = 200)
    private String shortDescription;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private Category category;

    @Column(name = "sub_category", length = 100)
    private String subCategory;

    // Pricing
    @Column(nullable = false, precision = 10, scale = 2)
    private BigDecimal price;

    @Column(name = "discounted_price", precision = 10, scale = 2)
    private BigDecimal discountedPrice;

    @Column(length = 3)
    private String currency = "INR";

    // Availability
    @Enumerated(EnumType.STRING)
    @Column(name = "availability_status")
    private AvailabilityStatus availabilityStatus = AvailabilityStatus.AVAILABLE;

    @Column(name = "stock_quantity")
    private Integer stockQuantity = 0;

    @Column(name = "low_stock_threshold")
    private Integer lowStockThreshold = 10;

    // Preparation
    @Column(name = "preparation_time", nullable = false)
    private Integer preparationTime = 10; // in minutes

    @Column(name = "serving_size", length = 100)
    private String servingSize;

    // Properties
    @Enumerated(EnumType.STRING)
    @Column(name = "spice_level")
    private SpiceLevel spiceLevel;

    @Column(name = "is_popular")
    private Boolean isPopular = false;

    @Column(name = "is_featured")
    private Boolean isFeatured = false;

    @Column(name = "is_new_item")
    private Boolean isNewItem = false;

    // Ratings
    @Column(precision = 3, scale = 2)
    private BigDecimal rating = BigDecimal.ZERO;

    @Column(name = "review_count")
    private Integer reviewCount = 0;

    // Metadata
    @Column(columnDefinition = "JSON")
    private String tags;

    @Column(name = "available_days", columnDefinition = "JSON")
    private String availableDays;

    @Column(name = "available_time_slots", columnDefinition = "JSON")
    private String availableTimeSlots;

    // User references
    @Column(name = "created_by", length = 36)
    private String createdBy;

    @Column(name = "updated_by", length = 36)
    private String updatedBy;

    @CreationTimestamp
    @Column(name = "created_at", nullable = false, updatable = false)
    private LocalDateTime createdAt;

    @UpdateTimestamp
    @Column(name = "updated_at")
    private LocalDateTime updatedAt;

    @Column(name = "deleted_at")
    private LocalDateTime deletedAt;

    // Relationships
    @OneToMany(mappedBy = "menuItem", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<MenuItemDietaryRestriction> dietaryRestrictions = new ArrayList<>();

    @OneToMany(mappedBy = "menuItem", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<MenuItemAllergen> allergens = new ArrayList<>();

    @OneToMany(mappedBy = "menuItem", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<MenuItemIngredient> ingredients = new ArrayList<>();

    @OneToOne(mappedBy = "menuItem", cascade = CascadeType.ALL, orphanRemoval = true)
    private MenuItemNutrition nutrition;

    @OneToMany(mappedBy = "menuItem", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<MenuItemCustomization> customizations = new ArrayList<>();

    @OneToMany(mappedBy = "menuItem", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<MenuItemReview> reviews = new ArrayList<>();

    @PrePersist
    public void prePersist() {
        if (id == null || id.isEmpty()) {
            id = java.util.UUID.randomUUID().toString();
        }
    }

    public enum Category {
        BREAKFAST, LUNCH, DINNER,
        BEVERAGES, SNACKS, DESSERTS,
        COMBO_MEALS, HEALTHY_OPTIONS, SPECIAL_OF_THE_DAY
    }

    public enum AvailabilityStatus {
        AVAILABLE, OUT_OF_STOCK,
        COMING_SOON, TEMPORARILY_UNAVAILABLE, SEASONAL
    }

    public enum SpiceLevel {
        NONE, MILD, MEDIUM, HOT, EXTRA_HOT
    }
}
