package com.example.map.config;

import com.example.map.security.JwtAuthenticationFilter;
import lombok.RequiredArgsConstructor;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.http.HttpMethod;
import org.springframework.security.config.annotation.method.configuration.EnableMethodSecurity;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;
import org.springframework.web.cors.CorsConfiguration;
import org.springframework.web.cors.CorsConfigurationSource;
import org.springframework.web.cors.UrlBasedCorsConfigurationSource;

import java.util.Arrays;
import java.util.List;

@Configuration
@EnableWebSecurity
@EnableMethodSecurity(prePostEnabled = true)
@RequiredArgsConstructor
public class SecurityConfig {

    private final JwtAuthenticationFilter jwtAuthenticationFilter;

    @Bean
    public SecurityFilterChain securityFilterChain(HttpSecurity http) throws Exception {
        http
                .cors(cors -> cors.configurationSource(corsConfigurationSource()))
                .csrf(csrf -> csrf.disable())
                .sessionManagement(session -> session.sessionCreationPolicy(SessionCreationPolicy.STATELESS))
                .authorizeHttpRequests(auth -> auth
                        // Public endpoints
                        .requestMatchers("/health", "/actuator/**").permitAll()
                        .requestMatchers(HttpMethod.GET, "/api/menu/categories/**").permitAll()
                        .requestMatchers(HttpMethod.GET, "/api/menu/items/**").permitAll()
                        
                        // Category endpoints - Admin only
                        .requestMatchers(HttpMethod.POST, "/api/menu/categories").hasRole("ADMIN")
                        .requestMatchers(HttpMethod.PUT, "/api/menu/categories/**").hasRole("ADMIN")
                        .requestMatchers(HttpMethod.DELETE, "/api/menu/categories/**").hasRole("ADMIN")
                        
                        // Menu item endpoints - Admin only
                        .requestMatchers(HttpMethod.POST, "/api/menu/items").hasRole("ADMIN")
                        .requestMatchers(HttpMethod.PUT, "/api/menu/items/**").hasRole("ADMIN")
                        .requestMatchers(HttpMethod.DELETE, "/api/menu/items/**").hasRole("ADMIN")
                        .requestMatchers(HttpMethod.PATCH, "/api/menu/items/**").hasRole("ADMIN")
                        
                        // Review endpoints - authenticated users
                        .requestMatchers(HttpMethod.POST, "/api/menu/items/*/reviews").authenticated()
                        .requestMatchers(HttpMethod.PUT, "/api/menu/items/*/reviews/*").authenticated()
                        .requestMatchers(HttpMethod.DELETE, "/api/menu/items/*/reviews/*").authenticated()
                        
                        // All other requests require authentication
                        .anyRequest().authenticated()
                )
                .addFilterBefore(jwtAuthenticationFilter, UsernamePasswordAuthenticationFilter.class);

        return http.build();
    }

    @Bean
    public CorsConfigurationSource corsConfigurationSource() {
        CorsConfiguration configuration = new CorsConfiguration();
        configuration.setAllowedOrigins(List.of("http://localhost:3000", "http://localhost:3001"));
        configuration.setAllowedMethods(Arrays.asList("GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"));
        configuration.setAllowedHeaders(Arrays.asList("*"));
        configuration.setAllowCredentials(true);
        
        UrlBasedCorsConfigurationSource source = new UrlBasedCorsConfigurationSource();
        source.registerCorsConfiguration("/**", configuration);
        return source;
    }
}
