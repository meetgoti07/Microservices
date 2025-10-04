package com.example.map.controller;

import com.example.map.dto.ApiResponse;
import com.example.map.dto.PageResponse;
import com.example.map.dto.ReviewRequest;
import com.example.map.dto.ReviewResponse;
import com.example.map.service.ReviewService;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/menu/items")
@RequiredArgsConstructor
@Slf4j
public class ReviewController {

    private final ReviewService reviewService;

    @GetMapping("/{menuItemId}/reviews")
    public ResponseEntity<ApiResponse<PageResponse<ReviewResponse>>> getReviewsByMenuItem(
            @PathVariable String menuItemId,
            @RequestParam(defaultValue = "0") Integer page,
            @RequestParam(defaultValue = "10") Integer size
    ) {
        log.info("GET /api/menu/items/{}/reviews - page: {}, size: {}", menuItemId, page, size);
        PageResponse<ReviewResponse> response = reviewService.getReviewsByMenuItem(menuItemId, page, size);
        return ResponseEntity.ok(ApiResponse.success(response));
    }

    @PostMapping("/{menuItemId}/reviews")
    public ResponseEntity<ApiResponse<ReviewResponse>> createReview(
            @PathVariable String menuItemId,
            @Valid @RequestBody ReviewRequest request,
            HttpServletRequest httpRequest
    ) {
        String userId = (String) httpRequest.getAttribute("userId");
        String userName = (String) httpRequest.getAttribute("userName");

        log.info("POST /api/menu/items/{}/reviews - by user: {}", menuItemId, userId);

        ReviewResponse response = reviewService.createReview(menuItemId, request, userId, userName);
        return ResponseEntity.status(HttpStatus.CREATED)
                .body(ApiResponse.success("Review created successfully", response));
    }

    @PutMapping("/{menuItemId}/reviews/{reviewId}")
    public ResponseEntity<ApiResponse<ReviewResponse>> updateReview(
            @PathVariable String menuItemId,
            @PathVariable String reviewId,
            @Valid @RequestBody ReviewRequest request,
            HttpServletRequest httpRequest
    ) {
        String userId = (String) httpRequest.getAttribute("userId");

        log.info("PUT /api/menu/items/{}/reviews/{} - by user: {}", menuItemId, reviewId, userId);

        ReviewResponse response = reviewService.updateReview(reviewId, request, userId);
        return ResponseEntity.ok(ApiResponse.success("Review updated successfully", response));
    }

    @DeleteMapping("/{menuItemId}/reviews/{reviewId}")
    public ResponseEntity<ApiResponse<Void>> deleteReview(
            @PathVariable String menuItemId,
            @PathVariable String reviewId,
            HttpServletRequest httpRequest
    ) {
        String userId = (String) httpRequest.getAttribute("userId");

        log.info("DELETE /api/menu/items/{}/reviews/{} - by user: {}", menuItemId, reviewId, userId);

        reviewService.deleteReview(reviewId, userId);
        return ResponseEntity.ok(ApiResponse.success("Review deleted successfully", null));
    }

    @GetMapping("/reviews/my-reviews")
    public ResponseEntity<ApiResponse<PageResponse<ReviewResponse>>> getMyReviews(
            HttpServletRequest httpRequest,
            @RequestParam(defaultValue = "0") Integer page,
            @RequestParam(defaultValue = "10") Integer size
    ) {
        String userId = (String) httpRequest.getAttribute("userId");

        log.info("GET /api/menu/items/reviews/my-reviews - user: {}", userId);

        PageResponse<ReviewResponse> response = reviewService.getUserReviews(userId, page, size);
        return ResponseEntity.ok(ApiResponse.success(response));
    }
}
