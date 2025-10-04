package com.example.map.dto;

import com.example.map.model.MenuItem;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class MenuItemListResponse {
    private String id;
    private String name;
    private String shortDescription;
    private MenuItem.Category category;
    private String subCategory;
    private BigDecimal price;
    private BigDecimal discountedPrice;
    private BigDecimal finalPrice;
    private String currency;
    private MenuItem.AvailabilityStatus availabilityStatus;
    private Boolean isAvailable;
    private Integer preparationTime;
    private MenuItem.SpiceLevel spiceLevel;
    private Boolean isPopular;
    private Boolean isFeatured;
    private Boolean isNewItem;
    private BigDecimal rating;
    private Integer reviewCount;
}
