package com.example.map.exception;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;
import java.util.Map;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class ErrorResponse {
    private Boolean success = false;
    private String message;
    private String error;
    private Integer status;
    private LocalDateTime timestamp;
    private String path;
    private Map<String, String> validationErrors;

    public ErrorResponse(String message, String error, Integer status, String path) {
        this.success = false;
        this.message = message;
        this.error = error;
        this.status = status;
        this.timestamp = LocalDateTime.now();
        this.path = path;
    }
}
