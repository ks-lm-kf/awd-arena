package handler

import (
	"strconv"

	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/internal/service"
	"github.com/gofiber/fiber/v3"
)

// DockerImageHandler handles Docker image management endpoints.
type DockerImageHandler struct {
	svc *service.DockerImageService
}

// NewDockerImageHandler creates a new DockerImageHandler.
func NewDockerImageHandler() *DockerImageHandler {
	return &DockerImageHandler{svc: &service.DockerImageService{}}
}

// List handles GET /api/v1/docker-images
func (h *DockerImageHandler) List(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	category := c.Query("category")
	difficulty := c.Query("difficulty")
	search := c.Query("search")

	result, err := h.svc.ListDockerImages(c.Context(), page, pageSize, category, difficulty, search)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": result})
}

// Get handles GET /api/v1/docker-images/:id
func (h *DockerImageHandler) Get(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid id"})
	}
	img, err := h.svc.GetDockerImage(c.Context(), uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "docker image not found"})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": img})
}

// Create handles POST /api/v1/docker-images
func (h *DockerImageHandler) Create(c fiber.Ctx) error {
	var img model.DockerImage
	if err := c.Bind().Body(&img); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}
	if img.Name == "" {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "name is required"})
	}
	if err := h.svc.CreateDockerImage(c.Context(), &img); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"code": 0, "message": "ok", "data": img})
}

// Update handles PUT /api/v1/docker-images/:id
func (h *DockerImageHandler) Update(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid id"})
	}
	var img model.DockerImage
	if err := c.Bind().Body(&img); err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid request body"})
	}
	if err := h.svc.UpdateDockerImage(c.Context(), uint(id), &img); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok"})
}

// Delete handles DELETE /api/v1/docker-images/:id
func (h *DockerImageHandler) Delete(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid id"})
	}
	if err := h.svc.DeleteDockerImage(c.Context(), uint(id)); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok"})
}

// Pull handles POST /api/v1/docker-images/:id/pull
func (h *DockerImageHandler) Pull(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": "invalid id"})
	}
	img, err := h.svc.GetDockerImage(c.Context(), uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"code": 404, "message": "docker image not found"})
	}
	output, err := h.svc.ImportDockerImage(c.Context(), img.Name, img.Tag)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": fiber.Map{"output": output}})
}

// HostList handles GET /api/v1/docker-images/host/list
func (h *DockerImageHandler) HostList(c fiber.Ctx) error {
	images, err := h.svc.ListDockerImagesFromHost(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": images})
}
