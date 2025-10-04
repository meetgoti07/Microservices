package com.example.map.model;

import com.fasterxml.jackson.annotation.JsonIgnore;
import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.ArrayList;
import java.util.List;

@Entity
@Table(name = "menu_item_customizations",
        uniqueConstraints = @UniqueConstraint(columnNames = {"menu_item_id", "customization_id"}))
@Data
@NoArgsConstructor
@AllArgsConstructor
public class MenuItemCustomization {

    @Id
    @Column(length = 36)
    private String id;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "menu_item_id", nullable = false)
    @JsonIgnore
    private MenuItem menuItem;

    @Column(name = "customization_id", nullable = false, length = 100)
    private String customizationId;

    @Column(nullable = false, length = 100)
    private String name;

    @Column(name = "is_required")
    private Boolean isRequired = false;

    @Column(name = "max_selections")
    private Integer maxSelections = 1;

    @Column(name = "display_order")
    private Integer displayOrder = 0;

    @OneToMany(mappedBy = "customization", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<MenuItemCustomizationOption> options = new ArrayList<>();

    @PrePersist
    public void prePersist() {
        if (id == null || id.isEmpty()) {
            id = java.util.UUID.randomUUID().toString();
        }
    }
}
