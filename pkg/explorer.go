package explorer

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/regclient/regclient/types/ref"
	"github.com/telday/container-registry-explorer/pkg/client"
)

func GetImageNames(registry string) []string {
	client := client.RegistryClient(context.TODO())
	repos, err := client.RepoList(context.Background(), registry)

	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Unable to find images")
		return []string{}
	}

	images, _ := repos.GetRepos()

	return images
}

func GetTags(imageName string) []string {
	client := client.RegistryClient(context.TODO())
	ref, _ := ref.New(imageName)
	tags, _ := client.TagList(context.Background(), ref)
	return tags.Tags
}

// GetImageDigest returns the digest (SHA) for a specific image:tag
func GetImageDigest(imageRef string) (string, error) {
	c := client.RegistryClient(context.TODO())
	r, err := ref.New(imageRef)
	if err != nil {
		return "", fmt.Errorf("invalid image reference: %w", err)
	}

	manifest, err := c.ManifestHead(context.Background(), r)
	if err != nil {
		return "", fmt.Errorf("failed to get manifest: %w", err)
	}

	return manifest.GetDescriptor().Digest.String(), nil
}

// PullImage pulls a specific image:tag using docker CLI
func PullImage(imageRef string) error {
	cmd := exec.Command("docker", "pull", imageRef)
	return cmd.Run()
}

// PullImageAsync pulls an image and returns output through channels
func PullImageAsync(imageRef string) (chan string, chan error) {
	outputChan := make(chan string, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(outputChan)
		defer close(errChan)

		cmd := exec.Command("docker", "pull", imageRef)
		output, err := cmd.CombinedOutput()
		if err != nil {
			errChan <- fmt.Errorf("pull failed: %w - %s", err, string(output))
			return
		}
		outputChan <- string(output)
	}()

	return outputChan, errChan
}
