package com.example.map.model;

import com.fasterxml.jackson.annotation.JsonIgnore;
import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.UpdateTimestamp;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Entity
@Table(name = "menu_item_nutrition")
@Data
@NoArgsConstructor
@AllArgsConstructor
public class MenuItemNutrition {

    @Id
    @Column(length = 36)
    private String id;

    @OneToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "menu_item_id", nullable = false, unique = true)
    @JsonIgnore
    private MenuItem menuItem;

    @Column(precision = 8, scale = 2)
    private BigDecimal calories;

    @Column(precision = 8, scale = 2)
    private BigDecimal protein; // grams

    @Column(precision = 8, scale = 2)
    private BigDecimal carbohydrates; // grams

    @Column(precision = 8, scale = 2)
    private BigDecimal fat; // grams

    @Column(precision = 8, scale = 2)
    private BigDecimal fiber; // grams

    @Column(precision = 8, scale = 2)
    private BigDecimal sodium; // mg

    @Column(precision = 8, scale = 2)
    private BigDecimal sugar; // grams

    @UpdateTimestamp
    @Column(name = "updated_at")
    private LocalDateTime updatedAt;

    @PrePersist
    public void prePersist() {
        if (id == null || id.isEmpty()) {
            id = java.util.UUID.randomUUID().toString();
        }
    }
}
