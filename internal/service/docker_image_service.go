package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/model"
	"gorm.io/gorm"
)

// DockerImageService handles Docker image management.
type DockerImageService struct{}

type DockerImageListResult struct {
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
	Items []model.DockerImage `json:"items"`
}

type HostImage struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	ImageID    string `json:"image_id"`
	Size       string `json:"size"`
	CreatedAt  string `json:"created_at"`
}

func (s *DockerImageService) ListDockerImages(ctx context.Context, page, pageSize int, category, difficulty, search string) (*DockerImageListResult, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := db.Model(&model.DockerImage{})
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	if search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	var items []model.DockerImage
	err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	if err != nil {
		return nil, err
	}

	return &DockerImageListResult{
		Total: total,
		Page:  page,
		Size:  pageSize,
		Items: items,
	}, nil
}

func (s *DockerImageService) GetDockerImage(ctx context.Context, id uint) (*model.DockerImage, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}
	var img model.DockerImage
	err := db.First(&img, id).Error
	return &img, err
}

func (s *DockerImageService) CreateDockerImage(ctx context.Context, img *model.DockerImage) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}
	if img.Status == "" {
		img.Status = "active"
	}
	if img.Tag == "" {
		img.Tag = "latest"
	}
	if img.Category == "" {
		img.Category = "general"
	}
	if img.Difficulty == "" {
		img.Difficulty = "medium"
	}
	if img.InitialScore == 0 {
		img.InitialScore = 100
	}
	if img.MemoryLimit == 0 {
		img.MemoryLimit = 256
	}
	if img.CPULimit == 0 {
		img.CPULimit = 1.0
	}
	return db.Create(img).Error
}

func (s *DockerImageService) UpdateDockerImage(ctx context.Context, id uint, img *model.DockerImage) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}
	return db.Model(&model.DockerImage{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":          img.Name,
		"tag":           img.Tag,
		"image_id":      img.ImageID,
		"description":   img.Description,
		"category":      img.Category,
		"difficulty":    img.Difficulty,
		"ports":         img.Ports,
		"memory_limit":  img.MemoryLimit,
		"cpu_limit":     img.CPULimit,
		"flag":          img.Flag,
		"initial_score": img.InitialScore,
		"status":        img.Status,
	}).Error
}

func (s *DockerImageService) DeleteDockerImage(ctx context.Context, id uint) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}
	return db.Delete(&model.DockerImage{}, id).Error
}

// ImportDockerImage pulls a Docker image from a registry.
func (s *DockerImageService) ImportDockerImage(ctx context.Context, name, tag string) (string, error) {
	if tag == "" {
		tag = "latest"
	}
	imageRef := name + ":" + tag

	cmd := exec.CommandContext(ctx, "docker", "pull", imageRef)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker pull failed: %s: %w", string(output), err)
	}
	return string(output), nil
}

// HostImageOutput represents the JSON output of `docker images --format`.
type HostImageOutput struct {
	Repository string `json:"Repository"`
	Tag        string `json:"Tag"`
	ID         string `json:"ID"`
	Size       string `json:"Size"`
	CreatedSince string `json:"CreatedSince"`
}

// ListDockerImagesFromHost returns images on the host machine.
func (s *DockerImageService) ListDockerImagesFromHost(ctx context.Context) ([]HostImage, error) {
	cmd := exec.CommandContext(ctx, "docker", "images", "--format", "{{json .}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("docker images failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var images []HostImage
	for _, line := range lines {
		if line == "" {
			continue
		}
		var hi HostImageOutput
		if err := json.Unmarshal([]byte(line), &hi); err != nil {
			continue
		}
		images = append(images, HostImage{
			Repository: hi.Repository,
			Tag:        hi.Tag,
			ImageID:    hi.ID,
			Size:       hi.Size,
			CreatedAt:  hi.CreatedSince,
		})
	}
	return images, nil
}

// Ensure DockerImageService implements a non-empty interface (avoids unused import).
var _ *gorm.DB = nil
