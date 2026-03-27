// Extended Docker image operations
package container

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/awd-platform/awd-arena/pkg/logger"
)

// BuildImageOptions holds options for building a Docker image
type BuildImageOptions struct {
	ContextPath string            // Path to build context directory
	Dockerfile  string            // Path to Dockerfile (relative to context)
	Tags        []string          // Image tags
	BuildArgs   map[string]*string // Build arguments
	NoCache     bool              // Disable cache
	Labels      map[string]string // Image labels
}

// ImageDetails holds detailed image information
type ImageDetails struct {
	ID           string
	RepoTags     []string
	RepoDigests  []string
	Created      string
	Size         int64
	Architecture string
	Os           string
	Author       string
	Config       *container.Config
}

// RemoveImage removes a Docker image from the host
func RemoveImage(cli *client.Client, ctx context.Context, imageID string, force bool) error {
	_, err := cli.ImageRemove(ctx, imageID, image.RemoveOptions{
		Force:         force,
		PruneChildren: false,
	})
	if err != nil {
		logger.Error("failed to remove image", "image", imageID, "error", err)
		return err
	}
	logger.Info("image removed", "image", imageID)
	return nil
}

// PushImage pushes a Docker image to a registry
func PushImage(cli *client.Client, ctx context.Context, ref string, authConfig registry.AuthConfig) error {
	authBytes, err := json.Marshal(authConfig)
	if err != nil {
		return fmt.Errorf("marshal auth config: %w", err)
	}
	authBase64 := base64.URLEncoding.EncodeToString(authBytes)
	
	reader, err := cli.ImagePush(ctx, ref, image.PushOptions{
		RegistryAuth: authBase64,
	})
	if err != nil {
		logger.Error("failed to push image", "ref", ref, "error", err)
		return err
	}
	defer reader.Close()
	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return err
	}
	logger.Info("image pushed", "ref", ref)
	return nil
}

// BuildImage builds a Docker image from a Dockerfile
func BuildImage(cli *client.Client, ctx context.Context, opts BuildImageOptions) (string, error) {
	buildCtx, relDockerfile, err := prepareBuildContext(opts.ContextPath, opts.Dockerfile)
	if err != nil {
		return "", fmt.Errorf("prepare build context: %w", err)
	}
	defer buildCtx.Close()

	resp, err := cli.ImageBuild(ctx, buildCtx, types.ImageBuildOptions{
		Dockerfile:  relDockerfile,
		Tags:        opts.Tags,
		BuildArgs:   opts.BuildArgs,
		Remove:      true,
		ForceRemove: true,
		NoCache:     opts.NoCache,
		Labels:      opts.Labels,
	})
	if err != nil {
		logger.Error("failed to build image", "tags", opts.Tags, "error", err)
		return "", fmt.Errorf("image build: %w", err)
	}
	defer resp.Body.Close()

	// Read build output to ensure build completes
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read build output: %w", err)
	}

	// Extract image ID from build output
	var imageID string
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		if strings.Contains(line, "successfully built") || strings.Contains(line, "writing image") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if (part == "built" || part == "sha256:") && i+1 < len(parts) {
					imageID = parts[i+1]
					if strings.HasPrefix(imageID, "sha256:") {
						imageID = strings.TrimPrefix(imageID, "sha256:")
					}
					break
				}
			}
		}
	}

	if imageID == "" && len(opts.Tags) > 0 {
		imageID = opts.Tags[0]
	}

	logger.Info("image built successfully", "image", imageID)
	return imageID, nil
}

// prepareBuildContext prepares the build context directory
func prepareBuildContext(contextPath, dockerfile string) (io.ReadCloser, string, error) {
	// If contextPath is empty, use current directory
	if contextPath == "" {
		contextPath = "."
	}

	// Create a tar archive of the build context
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	// Walk through the directory and add files to the tar
	err := filepath.Walk(contextPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if it's a directory
		if info.IsDir() {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		// Adjust header name to be relative to context path
		relPath, err := filepath.Rel(contextPath, path)
		if err != nil {
			return err
		}
		header.Name = relPath

		// Write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// Write file content
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tw, file)
		return err
	})

	if err != nil {
		return nil, "", err
	}

	return io.NopCloser(buf), dockerfile, nil
}

// GetImageDetails returns detailed information about a Docker image
func GetImageDetails(cli *client.Client, ctx context.Context, imageID string) (*ImageDetails, error) {
	inspect, _, err := cli.ImageInspectWithRaw(ctx, imageID)
	if err != nil {
		return nil, err
	}

	var size int64
	if inspect.Size > 0 {
		size = inspect.Size
	}

	return &ImageDetails{
		ID:           inspect.ID,
		RepoTags:     inspect.RepoTags,
		RepoDigests:  inspect.RepoDigests,
		Created:      inspect.Created,
		Size:         size,
		Architecture: inspect.Architecture,
		Os:          inspect.Os,
		Author:      inspect.Author,
		Config:      inspect.Config,
	}, nil
}
