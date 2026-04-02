package engine

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/awd-platform/awd-arena/pkg/logger"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func sanitizeFilePath(path string) string {
	reg := regexp.MustCompile(`^[a-zA-Z0-9_/\-.]+$`)
	if !reg.MatchString(path) {
		return "/flag"
	}
	if strings.Contains(path, "..") {
		return "/flag"
	}
	if !strings.HasPrefix(path, "/") {
		return "/flag"
	}
	return path
}

func shellEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "'", "'\\''")
	s = strings.ReplaceAll(s, "`", "\\`")
	s = strings.ReplaceAll(s, "$", `\$`)
	s = strings.ReplaceAll(s, "!", `\!`)
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

// FlagWriter writes flags to Docker containers.
type FlagWriter struct {
	dockerClient *client.Client
	defaultPath  string
	timeout      time.Duration
}

// NewFlagWriter creates a new flag writer.
func NewFlagWriter(dockerClient *client.Client) *FlagWriter {
	return &FlagWriter{
		dockerClient: dockerClient,
		defaultPath:  "/flag",
		timeout:      30 * time.Second,
	}
}

// WriteFlag writes a flag to a container.
func (fw *FlagWriter) WriteFlag(ctx context.Context, containerID, flag string, customPath ...string) error {
	if fw.dockerClient == nil {
		return fmt.Errorf("docker client is not initialized")
	}

	path := fw.defaultPath
	if len(customPath) > 0 && customPath[0] != "" {
		path = customPath[0]
	}

	ctx, cancel := context.WithTimeout(ctx, fw.timeout)
	defer cancel()

	safePath := sanitizeFilePath(path)
	safeFlag := shellEscape(flag)

	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd: []string{
			"sh", "-c",
			fmt.Sprintf("printf '%%s' '%s' > '%s' && chmod 600 '%s'",
				safeFlag,
				strings.ReplaceAll(safePath, "'", "'\\''"),
				strings.ReplaceAll(safePath, "'", "'\\''")),
		},
		User: "awd",
	}

	execResp, err := fw.dockerClient.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		logger.Error("failed to create exec instance", "container_id", containerID, "error", err)
		return fmt.Errorf("failed to create exec: %w", err)
	}

	// Attach to exec instance
	attachResp, err := fw.dockerClient.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{})
	if err != nil {
		logger.Error("failed to attach to exec", "exec_id", execResp.ID, "error", err)
		return fmt.Errorf("failed to attach to exec: %w", err)
	}
	defer attachResp.Close()

	// Start exec
	if err := fw.dockerClient.ContainerExecStart(ctx, execResp.ID, types.ExecStartCheck{}); err != nil {
		logger.Error("failed to start exec", "exec_id", execResp.ID, "error", err)
		return fmt.Errorf("failed to start exec: %w", err)
	}

	// Read output for debugging
	output, _ := io.ReadAll(attachResp.Reader)
	logger.Debug("exec output", "container_id", containerID, "output", string(output))

	// Inspect exec to check exit code
	inspectResp, err := fw.dockerClient.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		logger.Error("failed to inspect exec", "exec_id", execResp.ID, "error", err)
		return fmt.Errorf("failed to inspect exec: %w", err)
	}

	if inspectResp.ExitCode != 0 {
		return fmt.Errorf("exec failed with exit code %d", inspectResp.ExitCode)
	}

	logger.Info("flag written to container", "container_id", containerID, "path", path)
	return nil
}

// WriteFlagBatch writes flags to multiple containers in parallel.
func (fw *FlagWriter) WriteFlagBatch(ctx context.Context, flags map[string]string, customPath ...string) error {
	path := fw.defaultPath
	if len(customPath) > 0 && customPath[0] != "" {
		path = customPath[0]
	}

	errChan := make(chan error, len(flags))
	sem := make(chan struct{}, 10) // limit concurrency to 10

	for containerID, flag := range flags {
		go func(cid, f string) {
			sem <- struct{}{}
			defer func() { <-sem }()

			if err := fw.WriteFlag(ctx, cid, f, path); err != nil {
				errChan <- fmt.Errorf("container %s: %w", cid, err)
				return
			}
			errChan <- nil
		}(containerID, flag)
	}

	var errors []error
	for i := 0; i < len(flags); i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		logger.Error("batch flag write had errors", "error_count", len(errors))
		return fmt.Errorf("batch write failed with %d errors: %v", len(errors), errors)
	}

	logger.Info("batch flag write completed", "count", len(flags), "path", path)
	return nil
}

// ReadFlag reads a flag from a container (for verification).
func (fw *FlagWriter) ReadFlag(ctx context.Context, containerID string, customPath ...string) (string, error) {
	if fw.dockerClient == nil {
		return "", fmt.Errorf("docker client is not initialized")
	}

	path := fw.defaultPath
	if len(customPath) > 0 && customPath[0] != "" {
		path = customPath[0]
	}

	safePath := sanitizeFilePath(path)

	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"cat", safePath},
	}

	ctx, cancel := context.WithTimeout(ctx, fw.timeout)
	defer cancel()

	execResp, err := fw.dockerClient.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create exec: %w", err)
	}

	attachResp, err := fw.dockerClient.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{})
	if err != nil {
		return "", fmt.Errorf("failed to attach to exec: %w", err)
	}
	defer attachResp.Close()

	if err := fw.dockerClient.ContainerExecStart(ctx, execResp.ID, types.ExecStartCheck{}); err != nil {
		return "", fmt.Errorf("failed to start exec: %w", err)
	}

	output, err := io.ReadAll(attachResp.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to read output: %w", err)
	}

	return string(output), nil
}
