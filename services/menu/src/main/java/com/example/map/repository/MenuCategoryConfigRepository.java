package com.example.map.repository;

import com.example.map.model.MenuCategoryConfig;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface MenuCategoryConfigRepository extends JpaRepository<MenuCategoryConfig, String> {

    Optional<MenuCategoryConfig> findByCategory(String category);

    List<MenuCategoryConfig> findByIsActiveTrueOrderByDisplayOrderAsc();

    @Query("SELECT c FROM MenuCategoryConfig c ORDER BY c.displayOrder ASC")
    List<MenuCategoryConfig> findAllOrderedByDisplayOrder();

    Boolean existsByCategory(String category);
}
