package com.example.map.repository;

import com.example.map.model.MenuItem;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface MenuItemRepository extends JpaRepository<MenuItem, String>, JpaSpecificationExecutor<MenuItem> {

    // Find by ID excluding soft deleted
    @Query("SELECT m FROM MenuItem m WHERE m.id = :id AND m.deletedAt IS NULL")
    Optional<MenuItem> findByIdAndNotDeleted(@Param("id") String id);

    // Find all excluding soft deleted
    @Query("SELECT m FROM MenuItem m WHERE m.deletedAt IS NULL")
    Page<MenuItem> findAllNotDeleted(Pageable pageable);

    // Find by category
    @Query("SELECT m FROM MenuItem m WHERE m.category = :category AND m.deletedAt IS NULL")
    Page<MenuItem> findByCategoryAndNotDeleted(@Param("category") MenuItem.Category category, Pageable pageable);

    // Find by availability status
    @Query("SELECT m FROM MenuItem m WHERE m.availabilityStatus = :status AND m.deletedAt IS NULL")
    Page<MenuItem> findByAvailabilityStatusAndNotDeleted(@Param("status") MenuItem.AvailabilityStatus status, Pageable pageable);

    // Find by category and availability
    @Query("SELECT m FROM MenuItem m WHERE m.category = :category AND m.availabilityStatus = :status AND m.deletedAt IS NULL")
    Page<MenuItem> findByCategoryAndAvailabilityStatusAndNotDeleted(
            @Param("category") MenuItem.Category category,
            @Param("status") MenuItem.AvailabilityStatus status,
            Pageable pageable
    );

    // Search by name
    @Query("SELECT m FROM MenuItem m WHERE LOWER(m.name) LIKE LOWER(CONCAT('%', :query, '%')) AND m.deletedAt IS NULL")
    Page<MenuItem> searchByName(@Param("query") String query, Pageable pageable);

    // Find popular items
    @Query("SELECT m FROM MenuItem m WHERE m.isPopular = true AND m.deletedAt IS NULL")
    Page<MenuItem> findPopularItems(Pageable pageable);

    // Find featured items
    @Query("SELECT m FROM MenuItem m WHERE m.isFeatured = true AND m.deletedAt IS NULL")
    Page<MenuItem> findFeaturedItems(Pageable pageable);

    // Find new items
    @Query("SELECT m FROM MenuItem m WHERE m.isNewItem = true AND m.deletedAt IS NULL")
    Page<MenuItem> findNewItems(Pageable pageable);

    // Find by price range
    @Query("SELECT m FROM MenuItem m WHERE m.price BETWEEN :minPrice AND :maxPrice AND m.deletedAt IS NULL")
    Page<MenuItem> findByPriceRange(@Param("minPrice") java.math.BigDecimal minPrice,
                                     @Param("maxPrice") java.math.BigDecimal maxPrice,
                                     Pageable pageable);

    // Count by category
    @Query("SELECT COUNT(m) FROM MenuItem m WHERE m.category = :category AND m.deletedAt IS NULL")
    Long countByCategory(@Param("category") MenuItem.Category category);

    // Find items with low stock
    @Query("SELECT m FROM MenuItem m WHERE m.stockQuantity <= m.lowStockThreshold AND m.deletedAt IS NULL")
    List<MenuItem> findLowStockItems();

    // Check if name exists
    @Query("SELECT COUNT(m) > 0 FROM MenuItem m WHERE LOWER(m.name) = LOWER(:name) AND m.id <> :excludeId AND m.deletedAt IS NULL")
    Boolean existsByNameAndNotId(@Param("name") String name, @Param("excludeId") String excludeId);

    // Find items by dietary restriction
    @Query("SELECT DISTINCT m FROM MenuItem m JOIN m.dietaryRestrictions dr WHERE dr.restriction = :restriction AND m.deletedAt IS NULL")
    Page<MenuItem> findByDietaryRestriction(@Param("restriction") com.example.map.model.MenuItemDietaryRestriction.Restriction restriction, Pageable pageable);
}
