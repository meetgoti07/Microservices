package com.example.map.repository;

import com.example.map.model.MenuItemReview;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.math.BigDecimal;
import java.util.Optional;

@Repository
public interface MenuItemReviewRepository extends JpaRepository<MenuItemReview, String> {

    // Find reviews by menu item
    Page<MenuItemReview> findByMenuItemIdAndIsApprovedTrue(String menuItemId, Pageable pageable);

    // Find reviews by user
    Page<MenuItemReview> findByUserId(String userId, Pageable pageable);

    // Find review by user and menu item
    Optional<MenuItemReview> findByMenuItemIdAndUserId(String menuItemId, String userId);

    // Check if user has reviewed
    Boolean existsByMenuItemIdAndUserId(String menuItemId, String userId);

    // Get average rating for menu item
    @Query("SELECT AVG(r.rating) FROM MenuItemReview r WHERE r.menuItem.id = :menuItemId AND r.isApproved = true")
    BigDecimal getAverageRating(@Param("menuItemId") String menuItemId);

    // Count reviews for menu item
    @Query("SELECT COUNT(r) FROM MenuItemReview r WHERE r.menuItem.id = :menuItemId AND r.isApproved = true")
    Long countByMenuItemId(@Param("menuItemId") String menuItemId);

    // Find pending reviews (for admin)
    Page<MenuItemReview> findByIsApprovedFalse(Pageable pageable);

    // Find flagged reviews (for admin)
    Page<MenuItemReview> findByIsFlaggedTrue(Pageable pageable);
}
