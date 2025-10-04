package com.example.map.service;

import com.example.map.dto.PageResponse;
import com.example.map.dto.ReviewRequest;
import com.example.map.dto.ReviewResponse;
import com.example.map.mapper.MenuItemMapper;
import com.example.map.model.MenuItem;
import com.example.map.model.MenuItemReview;
import com.example.map.repository.MenuItemRepository;
import com.example.map.repository.MenuItemReviewRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.util.List;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
@Slf4j
public class ReviewService {

    private final MenuItemReviewRepository reviewRepository;
    private final MenuItemRepository menuItemRepository;
    private final MenuItemMapper menuItemMapper;

    @Transactional(readOnly = true)
    public PageResponse<ReviewResponse> getReviewsByMenuItem(String menuItemId, Integer page, Integer size) {
        Pageable pageable = PageRequest.of(page, size, Sort.by(Sort.Direction.DESC, "createdAt"));
        Page<MenuItemReview> reviewsPage = reviewRepository.findByMenuItemIdAndIsApprovedTrue(menuItemId, pageable);

        List<ReviewResponse> content = reviewsPage.getContent().stream()
                .map(menuItemMapper::toReviewResponse)
                .collect(Collectors.toList());

        return PageResponse.<ReviewResponse>builder()
                .content(content)
                .page(reviewsPage.getNumber())
                .size(reviewsPage.getSize())
                .totalElements(reviewsPage.getTotalElements())
                .totalPages(reviewsPage.getTotalPages())
                .first(reviewsPage.isFirst())
                .last(reviewsPage.isLast())
                .empty(reviewsPage.isEmpty())
                .build();
    }

    @Transactional
    public ReviewResponse createReview(String menuItemId, ReviewRequest request, String userId, String userName) {
        log.info("Creating review for menu item: {} by user: {}", menuItemId, userId);

        // Check if menu item exists
        MenuItem menuItem = menuItemRepository.findByIdAndNotDeleted(menuItemId)
                .orElseThrow(() -> new RuntimeException("Menu item not found with id: " + menuItemId));

        // Check if user has already reviewed this item
        if (reviewRepository.existsByMenuItemIdAndUserId(menuItemId, userId)) {
            throw new RuntimeException("You have already reviewed this item");
        }

        MenuItemReview review = new MenuItemReview();
        review.setMenuItem(menuItem);
        review.setUserId(userId);
        review.setOrderId(request.getOrderId());
        review.setRating(request.getRating());
        review.setComment(request.getComment());
        review.setUserName(userName);
        review.setIsApproved(true); // Auto-approve for now
        review.setIsVerifiedPurchase(request.getOrderId() != null);

        MenuItemReview savedReview = reviewRepository.save(review);

        // Update menu item rating
        updateMenuItemRating(menuItemId);

        log.info("Review created successfully with id: {}", savedReview.getId());
        return menuItemMapper.toReviewResponse(savedReview);
    }

    @Transactional
    public ReviewResponse updateReview(String reviewId, ReviewRequest request, String userId) {
        log.info("Updating review: {} by user: {}", reviewId, userId);

        MenuItemReview review = reviewRepository.findById(reviewId)
                .orElseThrow(() -> new RuntimeException("Review not found with id: " + reviewId));

        // Check if user owns this review
        if (!review.getUserId().equals(userId)) {
            throw new RuntimeException("You can only update your own reviews");
        }

        review.setRating(request.getRating());
        review.setComment(request.getComment());

        MenuItemReview updatedReview = reviewRepository.save(review);

        // Update menu item rating
        updateMenuItemRating(review.getMenuItem().getId());

        log.info("Review updated successfully: {}", reviewId);
        return menuItemMapper.toReviewResponse(updatedReview);
    }

    @Transactional
    public void deleteReview(String reviewId, String userId) {
        log.info("Deleting review: {} by user: {}", reviewId, userId);

        MenuItemReview review = reviewRepository.findById(reviewId)
                .orElseThrow(() -> new RuntimeException("Review not found with id: " + reviewId));

        // Check if user owns this review
        if (!review.getUserId().equals(userId)) {
            throw new RuntimeException("You can only delete your own reviews");
        }

        String menuItemId = review.getMenuItem().getId();
        reviewRepository.delete(review);

        // Update menu item rating
        updateMenuItemRating(menuItemId);

        log.info("Review deleted successfully: {}", reviewId);
    }

    @Transactional
    protected void updateMenuItemRating(String menuItemId) {
        BigDecimal avgRating = reviewRepository.getAverageRating(menuItemId);
        Long reviewCount = reviewRepository.countByMenuItemId(menuItemId);

        MenuItem menuItem = menuItemRepository.findByIdAndNotDeleted(menuItemId)
                .orElseThrow(() -> new RuntimeException("Menu item not found"));

        if (avgRating != null) {
            menuItem.setRating(avgRating.setScale(2, RoundingMode.HALF_UP));
        } else {
            menuItem.setRating(BigDecimal.ZERO);
        }

        menuItem.setReviewCount(reviewCount.intValue());
        menuItemRepository.save(menuItem);

        log.debug("Updated rating for menu item {}: {} ({} reviews)", menuItemId, avgRating, reviewCount);
    }

    @Transactional(readOnly = true)
    public PageResponse<ReviewResponse> getUserReviews(String userId, Integer page, Integer size) {
        Pageable pageable = PageRequest.of(page, size, Sort.by(Sort.Direction.DESC, "createdAt"));
        Page<MenuItemReview> reviewsPage = reviewRepository.findByUserId(userId, pageable);

        List<ReviewResponse> content = reviewsPage.getContent().stream()
                .map(menuItemMapper::toReviewResponse)
                .collect(Collectors.toList());

        return PageResponse.<ReviewResponse>builder()
                .content(content)
                .page(reviewsPage.getNumber())
                .size(reviewsPage.getSize())
                .totalElements(reviewsPage.getTotalElements())
                .totalPages(reviewsPage.getTotalPages())
                .first(reviewsPage.isFirst())
                .last(reviewsPage.isLast())
                .empty(reviewsPage.isEmpty())
                .build();
    }
}
