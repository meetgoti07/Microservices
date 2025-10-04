package com.example.map.service;

import com.example.map.dto.CategoryRequest;
import com.example.map.dto.CategoryResponse;
import com.example.map.model.MenuCategoryConfig;
import com.example.map.model.MenuItem;
import com.example.map.repository.MenuCategoryConfigRepository;
import com.example.map.repository.MenuItemRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.UUID;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
@Slf4j
public class CategoryService {

    private final MenuCategoryConfigRepository categoryConfigRepository;
    private final MenuItemRepository menuItemRepository;

    @Transactional(readOnly = true)
    public List<CategoryResponse> getAllCategories(Boolean activeOnly) {
        List<MenuCategoryConfig> categories;
        
        if (activeOnly != null && activeOnly) {
            categories = categoryConfigRepository.findByIsActiveTrueOrderByDisplayOrderAsc();
        } else {
            categories = categoryConfigRepository.findAllOrderedByDisplayOrder();
        }

        return categories.stream()
                .map(this::toCategoryResponse)
                .collect(Collectors.toList());
    }

    @Transactional(readOnly = true)
    public CategoryResponse getCategoryByName(String categoryName) {
        MenuCategoryConfig category = categoryConfigRepository.findByCategory(categoryName)
                .orElseThrow(() -> new RuntimeException("Category not found: " + categoryName));
        return toCategoryResponse(category);
    }

    @Transactional
    public CategoryResponse createCategory(CategoryRequest request) {
        // Check if category already exists
        if (categoryConfigRepository.existsByCategory(request.getCategory())) {
            throw new RuntimeException("Category already exists: " + request.getCategory());
        }

        MenuCategoryConfig category = new MenuCategoryConfig();
        category.setId(UUID.randomUUID().toString());
        category.setCategory(request.getCategory());
        category.setDisplayName(request.getDisplayName());
        category.setDescription(request.getDescription());
        category.setIcon(request.getIcon());
        category.setDisplayOrder(request.getDisplayOrder());
        category.setIsActive(request.getIsActive());
        category.setAvailableTimeStart(request.getAvailableTimeStart());
        category.setAvailableTimeEnd(request.getAvailableTimeEnd());

        MenuCategoryConfig saved = categoryConfigRepository.save(category);
        log.info("Created category: {}", saved.getCategory());
        
        return toCategoryResponse(saved);
    }

    @Transactional
    public CategoryResponse updateCategory(String id, CategoryRequest request) {
        MenuCategoryConfig category = categoryConfigRepository.findById(id)
                .orElseThrow(() -> new RuntimeException("Category not found with id: " + id));

        // Check if updating to a different category name that already exists
        if (!category.getCategory().equals(request.getCategory()) && 
            categoryConfigRepository.existsByCategory(request.getCategory())) {
            throw new RuntimeException("Category already exists: " + request.getCategory());
        }

        category.setCategory(request.getCategory());
        category.setDisplayName(request.getDisplayName());
        category.setDescription(request.getDescription());
        category.setIcon(request.getIcon());
        category.setDisplayOrder(request.getDisplayOrder());
        category.setIsActive(request.getIsActive());
        category.setAvailableTimeStart(request.getAvailableTimeStart());
        category.setAvailableTimeEnd(request.getAvailableTimeEnd());

        MenuCategoryConfig updated = categoryConfigRepository.save(category);
        log.info("Updated category: {}", updated.getCategory());
        
        return toCategoryResponse(updated);
    }

    @Transactional
    public void deleteCategory(String id) {
        MenuCategoryConfig category = categoryConfigRepository.findById(id)
                .orElseThrow(() -> new RuntimeException("Category not found with id: " + id));

        // Check if there are menu items using this category
        MenuItem.Category categoryEnum;
        try {
            categoryEnum = MenuItem.Category.valueOf(category.getCategory());
            Long itemCount = menuItemRepository.countByCategory(categoryEnum);
            if (itemCount > 0) {
                throw new RuntimeException("Cannot delete category with existing menu items. Count: " + itemCount);
            }
        } catch (IllegalArgumentException e) {
            // If category doesn't match enum, it's safe to delete
            log.warn("Category doesn't match enum, proceeding with deletion: {}", category.getCategory());
        }

        categoryConfigRepository.delete(category);
        log.info("Deleted category: {}", category.getCategory());
    }

    private CategoryResponse toCategoryResponse(MenuCategoryConfig category) {
        // Get enum value from string
        MenuItem.Category categoryEnum;
        try {
            categoryEnum = MenuItem.Category.valueOf(category.getCategory());
        } catch (IllegalArgumentException e) {
            log.warn("Invalid category enum: {}", category.getCategory());
            categoryEnum = null;
        }

        Long itemCount = categoryEnum != null ? menuItemRepository.countByCategory(categoryEnum) : 0L;

        return CategoryResponse.builder()
                .id(category.getId())
                .category(category.getCategory())
                .displayName(category.getDisplayName())
                .description(category.getDescription())
                .icon(category.getIcon())
                .displayOrder(category.getDisplayOrder())
                .isActive(category.getIsActive())
                .availableTimeStart(category.getAvailableTimeStart())
                .availableTimeEnd(category.getAvailableTimeEnd())
                .itemCount(itemCount)
                .build();
    }
}
