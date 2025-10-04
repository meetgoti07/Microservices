package com.example.map.model;

import com.fasterxml.jackson.annotation.JsonIgnore;
import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Entity
@Table(name = "menu_item_ingredients")
@Data
@NoArgsConstructor
@AllArgsConstructor
public class MenuItemIngredient {

    @Id
    @Column(length = 36)
    private String id;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "menu_item_id", nullable = false)
    @JsonIgnore
    private MenuItem menuItem;

    @Column(nullable = false, length = 200)
    private String ingredient;

    @Column(name = "display_order")
    private Integer displayOrder = 0;

    @PrePersist
    public void prePersist() {
        if (id == null || id.isEmpty()) {
            id = java.util.UUID.randomUUID().toString();
        }
    }
}
