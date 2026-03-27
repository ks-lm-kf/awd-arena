package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/awd-platform/awd-arena/internal/container"
	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
)

// RemoveImageFromHost removes a Docker image from the host machine
func (s *DockerImageService) RemoveImageFromHost(ctx context.Context, imageIDorName string, force bool) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("docker client: %w", err)
	}
	defer cli.Close()

	return container.RemoveImage(cli, ctx, imageIDorName, force)
}

// RemoveImageFromDBAndHost removes a Docker image from both database and host
func (s *DockerImageService) RemoveImageFromDBAndHost(ctx context.Context, id uint, force bool) error {
	// First get the image from database
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}

	var img model.DockerImage
	err := db.First(&img, id).Error
	if err != nil {
		return err
	}

	// Remove from host
	if img.ImageID != "" {
		err = s.RemoveImageFromHost(ctx, img.ImageID, force)
		if err != nil {
			// Log error but continue to delete from DB
			fmt.Printf("Warning: failed to remove image from host: %v\n", err)
		}
	}

	// Remove from database
	return db.Delete(&model.DockerImage{}, id).Error
}

// PushImageToRegistry pushes a Docker image to a registry
func (s *DockerImageService) PushImageToRegistry(ctx context.Context, imageRef string, authConfig registry.AuthConfig) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("docker client: %w", err)
	}
	defer cli.Close()

	return container.PushImage(cli, ctx, imageRef, authConfig)
}

// BuildImageFromDockerfile builds a Docker image from a Dockerfile
func (s *DockerImageService) BuildImageFromDockerfile(ctx context.Context, opts container.BuildImageOptions) (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", fmt.Errorf("docker client: %w", err)
	}
	defer cli.Close()

	return container.BuildImage(cli, ctx, opts)
}

// GetImageDetailsFromHost gets detailed information about a Docker image from the host
func (s *DockerImageService) GetImageDetailsFromHost(ctx context.Context, imageID string) (*container.ImageDetails, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	defer cli.Close()

	return container.GetImageDetails(cli, ctx, imageID)
}
