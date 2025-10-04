# Menu Service Dockerfile (Spring Boot + Maven)

# Build stage
FROM maven:3.9-eclipse-temurin-21-alpine AS builder

WORKDIR /app

# Copy pom.xml and download dependencies
COPY pom.xml ./
RUN mvn dependency:go-offline -B

# Copy source code
COPY src ./src

# Build the application
RUN mvn clean package -DskipTests -B

# Production stage
FROM eclipse-temurin:21-jre-alpine

WORKDIR /app

# Create non-root user
RUN addgroup -S spring && adduser -S spring -G spring

# Copy the JAR file from builder
COPY --from=builder /app/target/*.jar app.jar

# Change ownership
RUN chown spring:spring app.jar

# Switch to non-root user
USER spring:spring

# Expose ports (HTTP and gRPC)
EXPOSE 3002 50051

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=60s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3002/actuator/health || exit 1

# Start the application
ENTRYPOINT ["java", "-Djava.security.egd=file:/dev/./urandom", "-jar", "app.jar"]
