package com.example.map.dto;

import com.example.map.model.MenuItem;
import com.example.map.model.MenuItemDietaryRestriction;
import jakarta.validation.constraints.*;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.util.List;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class MenuItemRequest {

    @NotBlank(message = "Name is required")
    @Size(max = 200, message = "Name must not exceed 200 characters")
    private String name;

    private String description;

    @Size(max = 200, message = "Short description must not exceed 200 characters")
    private String shortDescription;

    @NotNull(message = "Category is required")
    private MenuItem.Category category;

    @Size(max = 100, message = "Sub category must not exceed 100 characters")
    private String subCategory;

    @NotNull(message = "Price is required")
    @DecimalMin(value = "0.0", inclusive = false, message = "Price must be greater than 0")
    private BigDecimal price;

    @DecimalMin(value = "0.0", inclusive = false, message = "Discounted price must be greater than 0")
    private BigDecimal discountedPrice;

    @Size(max = 3, message = "Currency must not exceed 3 characters")
    private String currency = "INR";

    private MenuItem.AvailabilityStatus availabilityStatus = MenuItem.AvailabilityStatus.AVAILABLE;

    @Min(value = 0, message = "Stock quantity must be non-negative")
    private Integer stockQuantity = 0;

    @Min(value = 0, message = "Low stock threshold must be non-negative")
    private Integer lowStockThreshold = 10;

    @NotNull(message = "Preparation time is required")
    @Min(value = 1, message = "Preparation time must be at least 1 minute")
    private Integer preparationTime = 10;

    @Size(max = 100, message = "Serving size must not exceed 100 characters")
    private String servingSize;

    private MenuItem.SpiceLevel spiceLevel;

    private Boolean isPopular = false;
    private Boolean isFeatured = false;
    private Boolean isNewItem = false;

    private List<String> tags;
    private List<String> availableDays;
    private List<TimeSlot> availableTimeSlots;

    // Related entities
    private List<MenuItemDietaryRestriction.Restriction> dietaryRestrictions;
    private List<String> allergens;
    private List<IngredientRequest> ingredients;
    private NutritionRequest nutrition;
    private List<CustomizationRequest> customizations;

    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class TimeSlot {
        private String start; // HH:MM
        private String end;   // HH:MM
    }

    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class IngredientRequest {
        @NotBlank(message = "Ingredient name is required")
        private String ingredient;
        private Integer displayOrder = 0;
    }

    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class NutritionRequest {
        private BigDecimal calories;
        private BigDecimal protein;
        private BigDecimal carbohydrates;
        private BigDecimal fat;
        private BigDecimal fiber;
        private BigDecimal sodium;
        private BigDecimal sugar;
    }

    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class CustomizationRequest {
        @NotBlank(message = "Customization ID is required")
        private String customizationId;
        
        @NotBlank(message = "Customization name is required")
        private String name;
        
        private Boolean isRequired = false;
        private Integer maxSelections = 1;
        private Integer displayOrder = 0;
        private List<CustomizationOptionRequest> options;
    }

    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class CustomizationOptionRequest {
        @NotBlank(message = "Option label is required")
        private String label;
        
        @NotBlank(message = "Option value is required")
        private String value;
        
        private BigDecimal priceModifier = BigDecimal.ZERO;
        private Integer displayOrder = 0;
    }
}
