package com.example.map.model;

import com.fasterxml.jackson.annotation.JsonIgnore;
import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.time.LocalDateTime;

@Entity
@Table(name = "menu_item_reviews")
@Data
@NoArgsConstructor
@AllArgsConstructor
public class MenuItemReview {

    @Id
    @Column(length = 36)
    private String id;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "menu_item_id", nullable = false)
    @JsonIgnore
    private MenuItem menuItem;

    @Column(name = "user_id", nullable = false, length = 36)
    private String userId;

    @Column(name = "order_id", length = 36)
    private String orderId;

    @Column(nullable = false)
    private Integer rating;

    @Column(columnDefinition = "TEXT")
    private String comment;

    // User info cached
    @Column(name = "user_name", length = 100)
    private String userName;

    @Column(name = "user_avatar", length = 500)
    private String userAvatar;

    // Moderation
    @Column(name = "is_verified_purchase")
    private Boolean isVerifiedPurchase = false;

    @Column(name = "is_approved")
    private Boolean isApproved = true;

    @Column(name = "is_flagged")
    private Boolean isFlagged = false;

    @Column(name = "moderated_by", length = 36)
    private String moderatedBy;

    @Column(name = "moderation_notes", columnDefinition = "TEXT")
    private String moderationNotes;

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
        if (rating != null && (rating < 1 || rating > 5)) {
            throw new IllegalArgumentException("Rating must be between 1 and 5");
        }
    }
}
