package com.example.map.dto;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ReviewResponse {
    private String id;
    private String menuItemId;
    private String userId;
    private String orderId;
    private Integer rating;
    private String comment;
    private String userName;
    private String userAvatar;
    private Boolean isVerifiedPurchase;
    private Boolean isApproved;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;
}
