package com.example.map.dto;

import com.example.map.model.MenuItem;
import com.example.map.model.MenuItemDietaryRestriction;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.List;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class MenuItemResponse {

    private String id;
    private String name;
    private String description;
    private String shortDescription;
    private MenuItem.Category category;
    private String subCategory;

    // Pricing
    private BigDecimal price;
    private BigDecimal discountedPrice;
    private String currency;
    private BigDecimal finalPrice; // Computed field

    // Availability
    private MenuItem.AvailabilityStatus availabilityStatus;
    private Integer stockQuantity;
    private Integer lowStockThreshold;
    private Boolean isAvailable; // Computed field

    // Preparation
    private Integer preparationTime;
    private String servingSize;

    // Properties
    private MenuItem.SpiceLevel spiceLevel;
    private Boolean isPopular;
    private Boolean isFeatured;
    private Boolean isNewItem;

    // Ratings
    private BigDecimal rating;
    private Integer reviewCount;

    // Metadata
    private List<String> tags;
    private List<String> availableDays;
    private List<TimeSlot> availableTimeSlots;

    // Related entities
    private List<MenuItemDietaryRestriction.Restriction> dietaryRestrictions;
    private List<String> allergens;
    private List<IngredientResponse> ingredients;
    private NutritionResponse nutrition;
    private List<CustomizationResponse> customizations;

    // Timestamps
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class TimeSlot {
        private String start;
        private String end;
    }

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class IngredientResponse {
        private String id;
        private String ingredient;
        private Integer displayOrder;
    }

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class NutritionResponse {
        private String id;
        private BigDecimal calories;
        private BigDecimal protein;
        private BigDecimal carbohydrates;
        private BigDecimal fat;
        private BigDecimal fiber;
        private BigDecimal sodium;
        private BigDecimal sugar;
    }

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class CustomizationResponse {
        private String id;
        private String customizationId;
        private String name;
        private Boolean isRequired;
        private Integer maxSelections;
        private Integer displayOrder;
        private List<CustomizationOptionResponse> options;
    }

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class CustomizationOptionResponse {
        private String id;
        private String label;
        private String value;
        private BigDecimal priceModifier;
        private Integer displayOrder;
    }
}
