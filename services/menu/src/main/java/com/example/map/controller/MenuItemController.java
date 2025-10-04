package com.example.map.controller;

import com.example.map.dto.*;
import com.example.map.model.MenuItem;
import com.example.map.model.MenuItemDietaryRestriction;
import com.example.map.service.MenuItemService;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.math.BigDecimal;

@RestController
@RequestMapping("/api/menu/items")
@RequiredArgsConstructor
@Slf4j
public class MenuItemController {

    private final MenuItemService menuItemService;

    @GetMapping
    public ResponseEntity<ApiResponse<PageResponse<MenuItemListResponse>>> getAllMenuItems(
            @RequestParam(required = false) MenuItem.Category category,
            @RequestParam(required = false) MenuItem.AvailabilityStatus availabilityStatus,
            @RequestParam(required = false) MenuItemDietaryRestriction.Restriction dietaryRestriction,
            @RequestParam(required = false) String search,
            @RequestParam(required = false) Boolean isPopular,
            @RequestParam(required = false) Boolean isFeatured,
            @RequestParam(required = false) Boolean isNewItem,
            @RequestParam(required = false) BigDecimal minPrice,
            @RequestParam(required = false) BigDecimal maxPrice,
            @RequestParam(defaultValue = "0") Integer page,
            @RequestParam(defaultValue = "20") Integer size,
            @RequestParam(defaultValue = "createdAt") String sortBy,
            @RequestParam(defaultValue = "desc") String sortDirection
    ) {
        log.info("GET /api/menu/items - category: {}, page: {}, size: {}", category, page, size);

        PageResponse<MenuItemListResponse> response = menuItemService.getAllMenuItems(
                category, availabilityStatus, dietaryRestriction, search,
                isPopular, isFeatured, isNewItem, minPrice, maxPrice,
                page, size, sortBy, sortDirection
        );

        return ResponseEntity.ok(ApiResponse.success(response));
    }

    @GetMapping("/{id}")
    public ResponseEntity<ApiResponse<MenuItemResponse>> getMenuItemById(@PathVariable String id) {
        log.info("GET /api/menu/items/{}", id);
        MenuItemResponse response = menuItemService.getMenuItemById(id);
        return ResponseEntity.ok(ApiResponse.success(response));
    }

    @PostMapping
    public ResponseEntity<ApiResponse<MenuItemResponse>> createMenuItem(
            @Valid @RequestBody MenuItemRequest request,
            HttpServletRequest httpRequest
    ) {
        String userId = (String) httpRequest.getAttribute("userId");
        String userRole = (String) httpRequest.getAttribute("userRole");

        log.info("POST /api/menu/items - by user: {} (role: {})", userId, userRole);
        log.info("Request attributes - userId: {}, userRole: {}", 
                httpRequest.getAttribute("userId"), 
                httpRequest.getAttribute("userRole"));

        // Check if user is admin
        if (userRole == null || !"admin".equalsIgnoreCase(userRole)) {
            log.warn("Access denied. User role: {} (required: admin)", userRole);
            return ResponseEntity.status(HttpStatus.FORBIDDEN)
                    .body(ApiResponse.error("Only admins can create menu items. Your role: " + 
                        (userRole != null ? userRole : "null")));
        }

        MenuItemResponse response = menuItemService.createMenuItem(request, userId);
        return ResponseEntity.status(HttpStatus.CREATED)
                .body(ApiResponse.success("Menu item created successfully", response));
    }

    @PutMapping("/{id}")
    public ResponseEntity<ApiResponse<MenuItemResponse>> updateMenuItem(
            @PathVariable String id,
            @Valid @RequestBody MenuItemRequest request,
            HttpServletRequest httpRequest
    ) {
        String userId = (String) httpRequest.getAttribute("userId");
        String userRole = (String) httpRequest.getAttribute("userRole");

        log.info("PUT /api/menu/items/{} - by user: {} (role: {})", id, userId, userRole);

        // Check if user is admin
        if (!"admin".equalsIgnoreCase(userRole)) {
            return ResponseEntity.status(HttpStatus.FORBIDDEN)
                    .body(ApiResponse.error("Only admins can update menu items"));
        }

        MenuItemResponse response = menuItemService.updateMenuItem(id, request, userId);
        return ResponseEntity.ok(ApiResponse.success("Menu item updated successfully", response));
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<ApiResponse<Void>> deleteMenuItem(
            @PathVariable String id,
            HttpServletRequest httpRequest
    ) {
        String userId = (String) httpRequest.getAttribute("userId");
        String userRole = (String) httpRequest.getAttribute("userRole");

        log.info("DELETE /api/menu/items/{} - by user: {} (role: {})", id, userId, userRole);

        // Check if user is admin
        if (!"admin".equalsIgnoreCase(userRole)) {
            return ResponseEntity.status(HttpStatus.FORBIDDEN)
                    .body(ApiResponse.error("Only admins can delete menu items"));
        }

        menuItemService.deleteMenuItem(id, userId);
        return ResponseEntity.ok(ApiResponse.success("Menu item deleted successfully", null));
    }

    @PatchMapping("/{id}/stock")
    public ResponseEntity<ApiResponse<MenuItemResponse>> updateStock(
            @PathVariable String id,
            @RequestParam Integer quantity,
            HttpServletRequest httpRequest
    ) {
        String userId = (String) httpRequest.getAttribute("userId");
        String userRole = (String) httpRequest.getAttribute("userRole");

        log.info("PATCH /api/menu/items/{}/stock - quantity: {}", id, quantity);

        // Check if user is admin
        if (!"admin".equalsIgnoreCase(userRole)) {
            return ResponseEntity.status(HttpStatus.FORBIDDEN)
                    .body(ApiResponse.error("Only admins can update stock"));
        }

        MenuItemResponse response = menuItemService.updateStock(id, quantity, userId);
        return ResponseEntity.ok(ApiResponse.success("Stock updated successfully", response));
    }

    @PatchMapping("/{id}/availability")
    public ResponseEntity<ApiResponse<MenuItemResponse>> updateAvailability(
            @PathVariable String id,
            @RequestParam MenuItem.AvailabilityStatus status,
            HttpServletRequest httpRequest
    ) {
        String userId = (String) httpRequest.getAttribute("userId");
        String userRole = (String) httpRequest.getAttribute("userRole");

        log.info("PATCH /api/menu/items/{}/availability - status: {}", id, status);

        // Check if user is admin
        if (!"admin".equalsIgnoreCase(userRole)) {
            return ResponseEntity.status(HttpStatus.FORBIDDEN)
                    .body(ApiResponse.error("Only admins can update availability"));
        }

        MenuItemResponse response = menuItemService.updateAvailability(id, status, userId);
        return ResponseEntity.ok(ApiResponse.success("Availability updated successfully", response));
    }
}
