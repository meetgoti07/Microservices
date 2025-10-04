package com.example.map.service;

import com.example.map.dto.*;
import com.example.map.mapper.MenuItemMapper;
import com.example.map.model.MenuItem;
import com.example.map.model.MenuItemDietaryRestriction;
import com.example.map.repository.MenuItemRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.data.jpa.domain.Specification;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.List;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
@Slf4j
public class MenuItemService {

    private final MenuItemRepository menuItemRepository;
    private final MenuItemMapper menuItemMapper;

    @Transactional(readOnly = true)
    public PageResponse<MenuItemListResponse> getAllMenuItems(
            MenuItem.Category category,
            MenuItem.AvailabilityStatus availabilityStatus,
            MenuItemDietaryRestriction.Restriction dietaryRestriction,
            String search,
            Boolean isPopular,
            Boolean isFeatured,
            Boolean isNewItem,
            BigDecimal minPrice,
            BigDecimal maxPrice,
            Integer page,
            Integer size,
            String sortBy,
            String sortDirection
    ) {
        Sort.Direction direction = Sort.Direction.fromString(sortDirection.toUpperCase());
        Pageable pageable = PageRequest.of(page, size, Sort.by(direction, sortBy));

        Page<MenuItem> menuItemsPage;

        // Build specification for complex queries
        Specification<MenuItem> spec = Specification.where(notDeleted());

        if (category != null) {
            spec = spec.and(hasCategory(category));
        }

        if (availabilityStatus != null) {
            spec = spec.and(hasAvailabilityStatus(availabilityStatus));
        }

        if (isPopular != null && isPopular) {
            spec = spec.and(isPopular());
        }

        if (isFeatured != null && isFeatured) {
            spec = spec.and(isFeatured());
        }

        if (isNewItem != null && isNewItem) {
            spec = spec.and(isNewItem());
        }

        if (search != null && !search.isEmpty()) {
            spec = spec.and(nameContains(search));
        }

        if (minPrice != null && maxPrice != null) {
            spec = spec.and(priceBetween(minPrice, maxPrice));
        }

        if (dietaryRestriction != null) {
            menuItemsPage = menuItemRepository.findByDietaryRestriction(dietaryRestriction, pageable);
        } else {
            menuItemsPage = menuItemRepository.findAll(spec, pageable);
        }

        List<MenuItemListResponse> content = menuItemsPage.getContent().stream()
                .map(menuItemMapper::toListResponse)
                .collect(Collectors.toList());

        return PageResponse.<MenuItemListResponse>builder()
                .content(content)
                .page(menuItemsPage.getNumber())
                .size(menuItemsPage.getSize())
                .totalElements(menuItemsPage.getTotalElements())
                .totalPages(menuItemsPage.getTotalPages())
                .first(menuItemsPage.isFirst())
                .last(menuItemsPage.isLast())
                .empty(menuItemsPage.isEmpty())
                .build();
    }

    @Transactional(readOnly = true)
    public MenuItemResponse getMenuItemById(String id) {
        MenuItem menuItem = menuItemRepository.findByIdAndNotDeleted(id)
                .orElseThrow(() -> new RuntimeException("Menu item not found with id: " + id));
        return menuItemMapper.toResponse(menuItem);
    }

    @Transactional
    public MenuItemResponse createMenuItem(MenuItemRequest request, String userId) {
        log.info("Creating menu item: {} by user: {}", request.getName(), userId);

        MenuItem menuItem = menuItemMapper.toEntity(request, userId);
        menuItem.setCreatedBy(userId);
        menuItem.setUpdatedBy(userId);

        MenuItem savedItem = menuItemRepository.save(menuItem);
        log.info("Menu item created successfully with id: {}", savedItem.getId());

        return menuItemMapper.toResponse(savedItem);
    }

    @Transactional
    public MenuItemResponse updateMenuItem(String id, MenuItemRequest request, String userId) {
        log.info("Updating menu item: {} by user: {}", id, userId);

        MenuItem menuItem = menuItemRepository.findByIdAndNotDeleted(id)
                .orElseThrow(() -> new RuntimeException("Menu item not found with id: " + id));

        menuItemMapper.updateEntityFromRequest(menuItem, request, userId);

        MenuItem updatedItem = menuItemRepository.save(menuItem);
        log.info("Menu item updated successfully: {}", id);

        return menuItemMapper.toResponse(updatedItem);
    }

    @Transactional
    public void deleteMenuItem(String id, String userId) {
        log.info("Deleting menu item: {} by user: {}", id, userId);

        MenuItem menuItem = menuItemRepository.findByIdAndNotDeleted(id)
                .orElseThrow(() -> new RuntimeException("Menu item not found with id: " + id));

        // Soft delete
        menuItem.setDeletedAt(LocalDateTime.now());
        menuItem.setUpdatedBy(userId);
        menuItemRepository.save(menuItem);

        log.info("Menu item soft deleted successfully: {}", id);
    }

    @Transactional
    public MenuItemResponse updateStock(String id, Integer quantity, String userId) {
        log.info("Updating stock for menu item: {} to quantity: {}", id, quantity);

        MenuItem menuItem = menuItemRepository.findByIdAndNotDeleted(id)
                .orElseThrow(() -> new RuntimeException("Menu item not found with id: " + id));

        menuItem.setStockQuantity(quantity);
        menuItem.setUpdatedBy(userId);

        // Auto-update availability status based on stock
        if (quantity == 0) {
            menuItem.setAvailabilityStatus(MenuItem.AvailabilityStatus.OUT_OF_STOCK);
        } else if (menuItem.getAvailabilityStatus() == MenuItem.AvailabilityStatus.OUT_OF_STOCK) {
            menuItem.setAvailabilityStatus(MenuItem.AvailabilityStatus.AVAILABLE);
        }

        MenuItem updatedItem = menuItemRepository.save(menuItem);
        log.info("Stock updated successfully for menu item: {}", id);

        return menuItemMapper.toResponse(updatedItem);
    }

    @Transactional
    public MenuItemResponse updateAvailability(String id, MenuItem.AvailabilityStatus status, String userId) {
        log.info("Updating availability for menu item: {} to status: {}", id, status);

        MenuItem menuItem = menuItemRepository.findByIdAndNotDeleted(id)
                .orElseThrow(() -> new RuntimeException("Menu item not found with id: " + id));

        menuItem.setAvailabilityStatus(status);
        menuItem.setUpdatedBy(userId);

        MenuItem updatedItem = menuItemRepository.save(menuItem);
        log.info("Availability updated successfully for menu item: {}", id);

        return menuItemMapper.toResponse(updatedItem);
    }

    @Transactional(readOnly = true)
    public List<MenuItem> findLowStockItems() {
        return menuItemRepository.findLowStockItems();
    }

    // Specification methods for complex queries
    private Specification<MenuItem> notDeleted() {
        return (root, query, cb) -> cb.isNull(root.get("deletedAt"));
    }

    private Specification<MenuItem> hasCategory(MenuItem.Category category) {
        return (root, query, cb) -> cb.equal(root.get("category"), category);
    }

    private Specification<MenuItem> hasAvailabilityStatus(MenuItem.AvailabilityStatus status) {
        return (root, query, cb) -> cb.equal(root.get("availabilityStatus"), status);
    }

    private Specification<MenuItem> isPopular() {
        return (root, query, cb) -> cb.isTrue(root.get("isPopular"));
    }

    private Specification<MenuItem> isFeatured() {
        return (root, query, cb) -> cb.isTrue(root.get("isFeatured"));
    }

    private Specification<MenuItem> isNewItem() {
        return (root, query, cb) -> cb.isTrue(root.get("isNewItem"));
    }

    private Specification<MenuItem> nameContains(String search) {
        return (root, query, cb) -> cb.like(cb.lower(root.get("name")), "%" + search.toLowerCase() + "%");
    }

    private Specification<MenuItem> priceBetween(BigDecimal minPrice, BigDecimal maxPrice) {
        return (root, query, cb) -> cb.between(root.get("price"), minPrice, maxPrice);
    }
}
