package handler

import (
	"strconv"

	"github.com/awd-platform/awd-arena/internal/container"
	"github.com/docker/docker/api/types/registry"
	"github.com/gofiber/fiber/v3"
)

// RemoveFromHost handles DELETE /api/v1/admin/images/host/:id
// Removes image from host machine only
func (h *DockerImageHandler) RemoveFromHost(c fiber.Ctx) error {
	imageIDorName := c.Params("id")
	if imageIDorName == "" {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "image id or name is required"})
	}

	force := c.Query("force", "false") == "true"

	err := h.svc.RemoveImageFromHost(c.Context(), imageIDorName, force)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}

	return c.JSON(fiber.Map{"code": 0, "message": "image removed from host"})
}

// RemoveFromDBAndHost handles DELETE /api/v1/admin/images/:id/complete
// Removes image from both database and host
func (h *DockerImageHandler) RemoveFromDBAndHost(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid id"})
	}

	force := c.Query("force", "false") == "true"

	err = h.svc.RemoveImageFromDBAndHost(c.Context(), uint(id), force)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}

	return c.JSON(fiber.Map{"code": 0, "message": "image removed completely"})
}

// PullImage handles POST /api/v1/admin/images/pull
// Pulls an image by name from registry
func (h *DockerImageHandler) PullImage(c fiber.Ctx) error {
	var req struct {
		Name string `json:"name"`
		Tag  string `json:"tag"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}
	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "name is required"})
	}

	output, err := h.svc.ImportDockerImage(c.Context(), req.Name, req.Tag)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}

	return c.JSON(fiber.Map{"code": 0, "message": "image pulled successfully", "data": fiber.Map{"output": output}})
}

// PushImage handles POST /api/v1/admin/images/push
// Pushes an image to registry
func (h *DockerImageHandler) PushImage(c fiber.Ctx) error {
	var req struct {
		ImageRef  string             `json:"image_ref"`
		AuthConfig registry.AuthConfig `json:"auth_config"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}
	if req.ImageRef == "" {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "image_ref is required"})
	}

	err := h.svc.PushImageToRegistry(c.Context(), req.ImageRef, req.AuthConfig)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}

	return c.JSON(fiber.Map{"code": 0, "message": "image pushed successfully"})
}

// BuildImage handles POST /api/v1/admin/images/build
// Builds an image from Dockerfile
func (h *DockerImageHandler) BuildImage(c fiber.Ctx) error {
	var req struct {
		ContextPath string            `json:"context_path"`
		Dockerfile  string            `json:"dockerfile"`
		Tags        []string          `json:"tags"`
		BuildArgs   map[string]string `json:"build_args"`
		NoCache     bool              `json:"no_cache"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}
	if len(req.Tags) == 0 {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "at least one tag is required"})
	}

	// Convert build args
	buildArgs := make(map[string]*string)
	for k, v := range req.BuildArgs {
		buildArgs[k] = &v
	}

	opts := container.BuildImageOptions{
		ContextPath: req.ContextPath,
		Dockerfile:  req.Dockerfile,
		Tags:        req.Tags,
		BuildArgs:   buildArgs,
		NoCache:     req.NoCache,
	}

	imageID, err := h.svc.BuildImageFromDockerfile(c.Context(), opts)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "image built successfully",
		"data": fiber.Map{
			"image_id": imageID,
		},
	})
}

// GetImageDetails handles GET /api/v1/admin/images/:id/details
// Gets detailed information about an image
func (h *DockerImageHandler) GetImageDetails(c fiber.Ctx) error {
	imageID := c.Params("id")
	if imageID == "" {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "image id is required"})
	}

	details, err := h.svc.GetImageDetailsFromHost(c.Context(), imageID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}

	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": details})
}
