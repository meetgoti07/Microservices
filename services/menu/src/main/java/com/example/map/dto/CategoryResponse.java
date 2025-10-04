package com.example.map.dto;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CategoryResponse {
    private String id;
    private String category;
    private String displayName;
    private String description;
    private String icon;
    private Integer displayOrder;
    private Boolean isActive;
    private String availableTimeStart;
    private String availableTimeEnd;
    private Long itemCount; // Number of items in this category
}
