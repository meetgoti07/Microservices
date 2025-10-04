package com.example.map.model;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.time.LocalDateTime;

@Entity
@Table(name = "menu_categories_config")
@Data
@NoArgsConstructor
@AllArgsConstructor
public class MenuCategoryConfig {

    @Id
    @Column(length = 36)
    private String id;

    @Column(unique = true, nullable = false, length = 100)
    private String category;

    @Column(name = "display_name", nullable = false, length = 200)
    private String displayName;

    @Column(columnDefinition = "TEXT")
    private String description;

    @Column(length = 200)
    private String icon;

    @Column(name = "display_order")
    private Integer displayOrder = 0;

    @Column(name = "is_active")
    private Boolean isActive = true;

    @Column(name = "available_time_start", length = 5)
    private String availableTimeStart; // HH:MM

    @Column(name = "available_time_end", length = 5)
    private String availableTimeEnd; // HH:MM

    @CreationTimestamp
    @Column(name = "created_at", nullable = false, updatable = false)
    private LocalDateTime createdAt;

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
