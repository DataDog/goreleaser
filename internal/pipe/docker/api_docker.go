package docker

import (
	"context"
	"fmt"
)

func init() {
	registerManifester(useDocker, dockerManifester{})

	registerImager(useDocker, dockerImager{})
	registerImager(useBuildx, dockerImager{
		buildx: true,
	})
	registerImager(useBuildPack, buildPackImager{})
}

type dockerManifester struct{}

func (m dockerManifester) Create(ctx context.Context, manifest string, images, flags []string) error {
	_ = runCommand(ctx, ".", "docker", "manifest", "rm", manifest)

	args := []string{"manifest", "create", manifest}
	args = append(args, images...)
	args = append(args, flags...)

	if err := runCommand(ctx, ".", "docker", args...); err != nil {
		return fmt.Errorf("failed to create %s: %w", manifest, err)
	}
	return nil
}

func (m dockerManifester) Push(ctx context.Context, manifest string, flags []string) error {
	args := []string{"manifest", "push", manifest}
	args = append(args, flags...)
	if err := runCommand(ctx, ".", "docker", args...); err != nil {
		return fmt.Errorf("failed to push %s: %w", manifest, err)
	}
	return nil
}

type dockerImager struct {
	buildx bool
}

func (i dockerImager) Push(ctx context.Context, image string, flags []string, exporter string) error {
	if err := runCommand(ctx, ".", "docker", "push", image); err != nil {
		return fmt.Errorf("failed to push %s: %w", image, err)
	}
	return nil
}

func (i dockerImager) Build(ctx context.Context, root string, images, flags []string, exporter string) error {
	if err := runCommand(ctx, root, "docker", i.buildCommand(images, flags, exporter)...); err != nil {
		return fmt.Errorf("failed to build %s: %w", images[0], err)
	}
	return nil
}

func (i dockerImager) buildCommand(images, flags []string, exporter string) []string {
	base := []string{"build", "."}
	if i.buildx {
		base = []string{"buildx", "build", "."}
		switch exporter {
		case "docker":
			base = append(base, "--output", "type=docker")
		case "tarDockerFormat":
			base = append(base, "--output", "type=docker,dest=/tmp/artifact.tar")
		}
	}
	for _, image := range images {
		base = append(base, "-t", image)
	}
	base = append(base, flags...)
	return base
}
