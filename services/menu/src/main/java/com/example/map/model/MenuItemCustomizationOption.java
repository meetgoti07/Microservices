package com.example.map.model;

import com.fasterxml.jackson.annotation.JsonIgnore;
import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;

@Entity
@Table(name = "menu_item_customization_options")
@Data
@NoArgsConstructor
@AllArgsConstructor
public class MenuItemCustomizationOption {

    @Id
    @Column(length = 36)
    private String id;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "customization_id", nullable = false)
    @JsonIgnore
    private MenuItemCustomization customization;

    @Column(nullable = false, length = 100)
    private String label;

    @Column(nullable = false, length = 100)
    private String value;

    @Column(name = "price_modifier", precision = 10, scale = 2)
    private BigDecimal priceModifier = BigDecimal.ZERO;

    @Column(name = "display_order")
    private Integer displayOrder = 0;

    @PrePersist
    public void prePersist() {
        if (id == null || id.isEmpty()) {
            id = java.util.UUID.randomUUID().toString();
        }
    }
}
