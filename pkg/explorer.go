package explorer

import (
	"context"
	"fmt"

	"github.com/regclient/regclient/types/ref"
	"github.com/telday/registry-explorer/pkg/client"
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
