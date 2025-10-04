package com.example.map.controller;

import com.example.map.dto.ApiResponse;
import com.example.map.dto.CategoryRequest;
import com.example.map.dto.CategoryResponse;
import com.example.map.service.CategoryService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/menu/categories")
@RequiredArgsConstructor
@Slf4j
public class CategoryController {

    private final CategoryService categoryService;

    @GetMapping
    public ResponseEntity<ApiResponse<List<CategoryResponse>>> getAllCategories(
            @RequestParam(required = false, defaultValue = "true") Boolean activeOnly
    ) {
        log.info("GET /api/menu/categories - activeOnly: {}", activeOnly);
        List<CategoryResponse> response = categoryService.getAllCategories(activeOnly);
        return ResponseEntity.ok(ApiResponse.success(response));
    }

    @GetMapping("/{categoryName}")
    public ResponseEntity<ApiResponse<CategoryResponse>> getCategoryByName(@PathVariable String categoryName) {
        log.info("GET /api/menu/categories/{}", categoryName);
        CategoryResponse response = categoryService.getCategoryByName(categoryName);
        return ResponseEntity.ok(ApiResponse.success(response));
    }

    @PostMapping
    @PreAuthorize("hasRole('ADMIN')")
    public ResponseEntity<ApiResponse<CategoryResponse>> createCategory(
            @Valid @RequestBody CategoryRequest request
    ) {
        log.info("POST /api/menu/categories - Creating category: {}", request.getCategory());
        CategoryResponse response = categoryService.createCategory(request);
        return ResponseEntity.status(HttpStatus.CREATED)
                .body(ApiResponse.success("Category created successfully", response));
    }

    @PutMapping("/{id}")
    @PreAuthorize("hasRole('ADMIN')")
    public ResponseEntity<ApiResponse<CategoryResponse>> updateCategory(
            @PathVariable String id,
            @Valid @RequestBody CategoryRequest request
    ) {
        log.info("PUT /api/menu/categories/{} - Updating category", id);
        CategoryResponse response = categoryService.updateCategory(id, request);
        return ResponseEntity.ok(ApiResponse.success("Category updated successfully", response));
    }

    @DeleteMapping("/{id}")
    @PreAuthorize("hasRole('ADMIN')")
    public ResponseEntity<ApiResponse<String>> deleteCategory(@PathVariable String id) {
        log.info("DELETE /api/menu/categories/{} - Deleting category", id);
        categoryService.deleteCategory(id);
        return ResponseEntity.ok(ApiResponse.success("Category deleted successfully", null));
    }
}
