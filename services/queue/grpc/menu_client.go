package grpc

import (
	"context"
	"fmt"
	"log"
	"time"

	"gin-quickstart/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// MenuClient wraps the gRPC connection to Menu Service
type MenuClient struct {
	conn   *grpc.ClientConn
	client MenuServiceClient
}

// MenuItem represents a menu item from Menu Service
type MenuItem struct {
	ID              string
	Name            string
	Category        string
	PreparationTime int // in minutes
	Price           float64
	IsAvailable     bool
}

// MenuServiceClient interface for gRPC calls
type MenuServiceClient interface {
	GetMenuItem(ctx context.Context, itemID string) (*MenuItem, error)
	GetMenuItems(ctx context.Context, itemIDs []string) ([]*MenuItem, error)
	GetAveragePreparationTime(ctx context.Context, itemIDs []string) (int, error)
}

type menuServiceClient struct {
	// This will be replaced with actual gRPC client when proto is available
}

func NewMenuClient(cfg *config.Config) (*MenuClient, error) {
	address := fmt.Sprintf("%s:%s", cfg.MenuServiceHost, cfg.MenuServicePort)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create gRPC connection
	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Printf("Warning: Failed to connect to Menu Service: %v", err)
		// Return mock client for development
		return &MenuClient{
			conn:   nil,
			client: &mockMenuClient{},
		}, nil
	}

	log.Printf("Connected to Menu Service at %s", address)

	// Initialize actual gRPC client
	// TODO: Replace with generated proto client when available
	// client := pb.NewMenuServiceClient(conn)

	return &MenuClient{
		conn:   conn,
		client: &mockMenuClient{},
	}, nil
}

func (mc *MenuClient) Close() error {
	if mc.conn != nil {
		return mc.conn.Close()
	}
	return nil
}

func (mc *MenuClient) GetMenuItem(ctx context.Context, itemID string) (*MenuItem, error) {
	return mc.client.GetMenuItem(ctx, itemID)
}

func (mc *MenuClient) GetMenuItems(ctx context.Context, itemIDs []string) ([]*MenuItem, error) {
	return mc.client.GetMenuItems(ctx, itemIDs)
}

func (mc *MenuClient) GetAveragePreparationTime(ctx context.Context, itemIDs []string) (int, error) {
	return mc.client.GetAveragePreparationTime(ctx, itemIDs)
}

// Mock implementation for development
type mockMenuClient struct{}

func (m *mockMenuClient) GetMenuItem(ctx context.Context, itemID string) (*MenuItem, error) {
	return &MenuItem{
		ID:              itemID,
		Name:            "Sample Item",
		Category:        "Main Course",
		PreparationTime: 10,
		Price:           9.99,
		IsAvailable:     true,
	}, nil
}

func (m *mockMenuClient) GetMenuItems(ctx context.Context, itemIDs []string) ([]*MenuItem, error) {
	items := make([]*MenuItem, len(itemIDs))
	for i, id := range itemIDs {
		items[i] = &MenuItem{
			ID:              id,
			Name:            fmt.Sprintf("Item %s", id),
			PreparationTime: 10,
			Price:           9.99,
			IsAvailable:     true,
		}
	}
	return items, nil
}

func (m *mockMenuClient) GetAveragePreparationTime(ctx context.Context, itemIDs []string) (int, error) {
	// Mock: return 10 minutes average
	return 10, nil
}
