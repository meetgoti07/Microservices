package com.example.map.mapper;

import com.example.map.dto.*;
import com.example.map.model.*;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.stereotype.Component;

import java.math.BigDecimal;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.stream.Collectors;

@Component
public class MenuItemMapper {

    private final ObjectMapper objectMapper = new ObjectMapper();

    public MenuItemResponse toResponse(MenuItem menuItem) {
        if (menuItem == null) return null;

        MenuItemResponse response = MenuItemResponse.builder()
                .id(menuItem.getId())
                .name(menuItem.getName())
                .description(menuItem.getDescription())
                .shortDescription(menuItem.getShortDescription())
                .category(menuItem.getCategory())
                .subCategory(menuItem.getSubCategory())
                .price(menuItem.getPrice())
                .discountedPrice(menuItem.getDiscountedPrice())
                .currency(menuItem.getCurrency())
                .finalPrice(calculateFinalPrice(menuItem))
                .availabilityStatus(menuItem.getAvailabilityStatus())
                .stockQuantity(menuItem.getStockQuantity())
                .lowStockThreshold(menuItem.getLowStockThreshold())
                .isAvailable(isAvailable(menuItem))
                .preparationTime(menuItem.getPreparationTime())
                .servingSize(menuItem.getServingSize())
                .spiceLevel(menuItem.getSpiceLevel())
                .isPopular(menuItem.getIsPopular())
                .isFeatured(menuItem.getIsFeatured())
                .isNewItem(menuItem.getIsNewItem())
                .rating(menuItem.getRating())
                .reviewCount(menuItem.getReviewCount())
                .tags(parseJsonArray(menuItem.getTags()))
                .availableDays(parseJsonArray(menuItem.getAvailableDays()))
                .availableTimeSlots(parseTimeSlots(menuItem.getAvailableTimeSlots()))
                .createdAt(menuItem.getCreatedAt())
                .updatedAt(menuItem.getUpdatedAt())
                .build();

        // Map related entities
        if (menuItem.getDietaryRestrictions() != null) {
            response.setDietaryRestrictions(
                    menuItem.getDietaryRestrictions().stream()
                            .map(MenuItemDietaryRestriction::getRestriction)
                            .collect(Collectors.toList())
            );
        }

        if (menuItem.getAllergens() != null) {
            response.setAllergens(
                    menuItem.getAllergens().stream()
                            .map(MenuItemAllergen::getAllergen)
                            .collect(Collectors.toList())
            );
        }

        if (menuItem.getIngredients() != null) {
            response.setIngredients(
                    menuItem.getIngredients().stream()
                            .map(this::toIngredientResponse)
                            .collect(Collectors.toList())
            );
        }

        if (menuItem.getNutrition() != null) {
            response.setNutrition(toNutritionResponse(menuItem.getNutrition()));
        }

        if (menuItem.getCustomizations() != null) {
            response.setCustomizations(
                    menuItem.getCustomizations().stream()
                            .map(this::toCustomizationResponse)
                            .collect(Collectors.toList())
            );
        }

        return response;
    }

    public MenuItemListResponse toListResponse(MenuItem menuItem) {
        if (menuItem == null) return null;

        return MenuItemListResponse.builder()
                .id(menuItem.getId())
                .name(menuItem.getName())
                .shortDescription(menuItem.getShortDescription())
                .category(menuItem.getCategory())
                .subCategory(menuItem.getSubCategory())
                .price(menuItem.getPrice())
                .discountedPrice(menuItem.getDiscountedPrice())
                .finalPrice(calculateFinalPrice(menuItem))
                .currency(menuItem.getCurrency())
                .availabilityStatus(menuItem.getAvailabilityStatus())
                .isAvailable(isAvailable(menuItem))
                .preparationTime(menuItem.getPreparationTime())
                .spiceLevel(menuItem.getSpiceLevel())
                .isPopular(menuItem.getIsPopular())
                .isFeatured(menuItem.getIsFeatured())
                .isNewItem(menuItem.getIsNewItem())
                .rating(menuItem.getRating())
                .reviewCount(menuItem.getReviewCount())
                .build();
    }

    public MenuItem toEntity(MenuItemRequest request, String userId) {
        MenuItem menuItem = new MenuItem();
        updateEntityFromRequest(menuItem, request, userId);
        return menuItem;
    }

    public void updateEntityFromRequest(MenuItem menuItem, MenuItemRequest request, String userId) {
        menuItem.setName(request.getName());
        menuItem.setDescription(request.getDescription());
        menuItem.setShortDescription(request.getShortDescription());
        menuItem.setCategory(request.getCategory());
        menuItem.setSubCategory(request.getSubCategory());
        menuItem.setPrice(request.getPrice());
        menuItem.setDiscountedPrice(request.getDiscountedPrice());
        menuItem.setCurrency(request.getCurrency());
        menuItem.setAvailabilityStatus(request.getAvailabilityStatus());
        menuItem.setStockQuantity(request.getStockQuantity());
        menuItem.setLowStockThreshold(request.getLowStockThreshold());
        menuItem.setPreparationTime(request.getPreparationTime());
        menuItem.setServingSize(request.getServingSize());
        menuItem.setSpiceLevel(request.getSpiceLevel());
        menuItem.setIsPopular(request.getIsPopular());
        menuItem.setIsFeatured(request.getIsFeatured());
        menuItem.setIsNewItem(request.getIsNewItem());

        // JSON fields
        menuItem.setTags(toJsonArray(request.getTags()));
        menuItem.setAvailableDays(toJsonArray(request.getAvailableDays()));
        menuItem.setAvailableTimeSlots(toJsonTimeSlots(request.getAvailableTimeSlots()));

        menuItem.setUpdatedBy(userId);

        // Handle related entities
        updateDietaryRestrictions(menuItem, request.getDietaryRestrictions());
        updateAllergens(menuItem, request.getAllergens());
        updateIngredients(menuItem, request.getIngredients());
        updateNutrition(menuItem, request.getNutrition());
        updateCustomizations(menuItem, request.getCustomizations());
    }

    // Helper methods for related entities
    private void updateDietaryRestrictions(MenuItem menuItem, List<MenuItemDietaryRestriction.Restriction> restrictions) {
        if (restrictions == null) return;

        menuItem.getDietaryRestrictions().clear();
        for (MenuItemDietaryRestriction.Restriction restriction : restrictions) {
            MenuItemDietaryRestriction dr = new MenuItemDietaryRestriction();
            dr.setMenuItem(menuItem);
            dr.setRestriction(restriction);
            menuItem.getDietaryRestrictions().add(dr);
        }
    }

    private void updateAllergens(MenuItem menuItem, List<String> allergens) {
        if (allergens == null) return;

        menuItem.getAllergens().clear();
        for (String allergen : allergens) {
            MenuItemAllergen a = new MenuItemAllergen();
            a.setMenuItem(menuItem);
            a.setAllergen(allergen);
            menuItem.getAllergens().add(a);
        }
    }

    private void updateIngredients(MenuItem menuItem, List<MenuItemRequest.IngredientRequest> ingredients) {
        if (ingredients == null) return;

        menuItem.getIngredients().clear();
        for (MenuItemRequest.IngredientRequest ingredient : ingredients) {
            MenuItemIngredient i = new MenuItemIngredient();
            i.setMenuItem(menuItem);
            i.setIngredient(ingredient.getIngredient());
            i.setDisplayOrder(ingredient.getDisplayOrder());
            menuItem.getIngredients().add(i);
        }
    }

    private void updateNutrition(MenuItem menuItem, MenuItemRequest.NutritionRequest nutrition) {
        if (nutrition == null) return;

        MenuItemNutrition n = menuItem.getNutrition();
        if (n == null) {
            n = new MenuItemNutrition();
            n.setMenuItem(menuItem);
            menuItem.setNutrition(n);
        }

        n.setCalories(nutrition.getCalories());
        n.setProtein(nutrition.getProtein());
        n.setCarbohydrates(nutrition.getCarbohydrates());
        n.setFat(nutrition.getFat());
        n.setFiber(nutrition.getFiber());
        n.setSodium(nutrition.getSodium());
        n.setSugar(nutrition.getSugar());
    }

    private void updateCustomizations(MenuItem menuItem, List<MenuItemRequest.CustomizationRequest> customizations) {
        if (customizations == null) return;

        menuItem.getCustomizations().clear();
        for (MenuItemRequest.CustomizationRequest customization : customizations) {
            MenuItemCustomization c = new MenuItemCustomization();
            c.setMenuItem(menuItem);
            c.setCustomizationId(customization.getCustomizationId());
            c.setName(customization.getName());
            c.setIsRequired(customization.getIsRequired());
            c.setMaxSelections(customization.getMaxSelections());
            c.setDisplayOrder(customization.getDisplayOrder());

            if (customization.getOptions() != null) {
                for (MenuItemRequest.CustomizationOptionRequest option : customization.getOptions()) {
                    MenuItemCustomizationOption o = new MenuItemCustomizationOption();
                    o.setCustomization(c);
                    o.setLabel(option.getLabel());
                    o.setValue(option.getValue());
                    o.setPriceModifier(option.getPriceModifier());
                    o.setDisplayOrder(option.getDisplayOrder());
                    c.getOptions().add(o);
                }
            }

            menuItem.getCustomizations().add(c);
        }
    }

    // Response mappers for nested entities
    private MenuItemResponse.IngredientResponse toIngredientResponse(MenuItemIngredient ingredient) {
        return MenuItemResponse.IngredientResponse.builder()
                .id(ingredient.getId())
                .ingredient(ingredient.getIngredient())
                .displayOrder(ingredient.getDisplayOrder())
                .build();
    }

    private MenuItemResponse.NutritionResponse toNutritionResponse(MenuItemNutrition nutrition) {
        return MenuItemResponse.NutritionResponse.builder()
                .id(nutrition.getId())
                .calories(nutrition.getCalories())
                .protein(nutrition.getProtein())
                .carbohydrates(nutrition.getCarbohydrates())
                .fat(nutrition.getFat())
                .fiber(nutrition.getFiber())
                .sodium(nutrition.getSodium())
                .sugar(nutrition.getSugar())
                .build();
    }

    private MenuItemResponse.CustomizationResponse toCustomizationResponse(MenuItemCustomization customization) {
        MenuItemResponse.CustomizationResponse response = MenuItemResponse.CustomizationResponse.builder()
                .id(customization.getId())
                .customizationId(customization.getCustomizationId())
                .name(customization.getName())
                .isRequired(customization.getIsRequired())
                .maxSelections(customization.getMaxSelections())
                .displayOrder(customization.getDisplayOrder())
                .build();

        if (customization.getOptions() != null) {
            response.setOptions(
                    customization.getOptions().stream()
                            .map(this::toCustomizationOptionResponse)
                            .collect(Collectors.toList())
            );
        }

        return response;
    }

    private MenuItemResponse.CustomizationOptionResponse toCustomizationOptionResponse(MenuItemCustomizationOption option) {
        return MenuItemResponse.CustomizationOptionResponse.builder()
                .id(option.getId())
                .label(option.getLabel())
                .value(option.getValue())
                .priceModifier(option.getPriceModifier())
                .displayOrder(option.getDisplayOrder())
                .build();
    }

    // Review mappers
    public ReviewResponse toReviewResponse(MenuItemReview review) {
        return ReviewResponse.builder()
                .id(review.getId())
                .menuItemId(review.getMenuItem().getId())
                .userId(review.getUserId())
                .orderId(review.getOrderId())
                .rating(review.getRating())
                .comment(review.getComment())
                .userName(review.getUserName())
                .userAvatar(review.getUserAvatar())
                .isVerifiedPurchase(review.getIsVerifiedPurchase())
                .isApproved(review.getIsApproved())
                .createdAt(review.getCreatedAt())
                .updatedAt(review.getUpdatedAt())
                .build();
    }

    // Helper methods
    private BigDecimal calculateFinalPrice(MenuItem menuItem) {
        if (menuItem.getDiscountedPrice() != null && menuItem.getDiscountedPrice().compareTo(BigDecimal.ZERO) > 0) {
            return menuItem.getDiscountedPrice();
        }
        return menuItem.getPrice();
    }

    private Boolean isAvailable(MenuItem menuItem) {
        return menuItem.getAvailabilityStatus() == MenuItem.AvailabilityStatus.AVAILABLE
                && menuItem.getStockQuantity() > 0;
    }

    private List<String> parseJsonArray(String json) {
        if (json == null || json.isEmpty()) return Collections.emptyList();
        try {
            return objectMapper.readValue(json, new TypeReference<List<String>>() {});
        } catch (JsonProcessingException e) {
            return Collections.emptyList();
        }
    }

    private List<MenuItemResponse.TimeSlot> parseTimeSlots(String json) {
        if (json == null || json.isEmpty()) return Collections.emptyList();
        try {
            return objectMapper.readValue(json, new TypeReference<List<MenuItemResponse.TimeSlot>>() {});
        } catch (JsonProcessingException e) {
            return Collections.emptyList();
        }
    }

    private String toJsonArray(List<String> list) {
        if (list == null || list.isEmpty()) return null;
        try {
            return objectMapper.writeValueAsString(list);
        } catch (JsonProcessingException e) {
            return null;
        }
    }

    private String toJsonTimeSlots(List<MenuItemRequest.TimeSlot> timeSlots) {
        if (timeSlots == null || timeSlots.isEmpty()) return null;
        try {
            return objectMapper.writeValueAsString(timeSlots);
        } catch (JsonProcessingException e) {
            return null;
        }
    }
}
