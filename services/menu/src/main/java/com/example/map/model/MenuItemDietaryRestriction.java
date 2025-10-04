package com.example.map.model;

import com.fasterxml.jackson.annotation.JsonIgnore;
import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Entity
@Table(name = "menu_item_dietary_restrictions",
        uniqueConstraints = @UniqueConstraint(columnNames = {"menu_item_id", "restriction"}))
@Data
@NoArgsConstructor
@AllArgsConstructor
public class MenuItemDietaryRestriction {

    @Id
    @Column(length = 36)
    private String id;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "menu_item_id", nullable = false)
    @JsonIgnore
    private MenuItem menuItem;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private Restriction restriction;

    @PrePersist
    public void prePersist() {
        if (id == null || id.isEmpty()) {
            id = java.util.UUID.randomUUID().toString();
        }
    }

    public enum Restriction {
        VEGETARIAN, VEGAN, GLUTEN_FREE,
        NUT_FREE, DAIRY_FREE, HALAL,
        JAIN, KETO, LOW_CARB, HIGH_PROTEIN
    }
}
