package com.example.map.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CategoryRequest {
    
    @NotBlank(message = "Category name is required")
    private String category;
    
    @NotBlank(message = "Display name is required")
    private String displayName;
    
    private String description;
    
    private String icon;
    
    @NotNull(message = "Display order is required")
    private Integer displayOrder;
    
    @NotNull(message = "Active status is required")
    private Boolean isActive;
    
    private String availableTimeStart; // HH:MM format
    
    private String availableTimeEnd; // HH:MM format
}
