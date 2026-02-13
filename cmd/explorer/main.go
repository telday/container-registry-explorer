package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/telday/container-registry-explorer/internal"
	"github.com/urfave/cli/v3"
)

var registryArg string = "registry"

func main() {
	cmd := &cli.Command{
		Name:      "explorer",
		Usage:     "Explore a private docker registry",
		ArgsUsage: "registry-name",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name: registryArg,
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			registry := cmd.StringArg(registryArg)
			if registry == "" {
				return fmt.Errorf("The name of a registry must be provided.")
			} else if registry == "docker.io" {
				return fmt.Errorf("Provided registry cannot be default docker.")
			}
			explorer := internal.NewExplorerApp(registry)
			return explorer.Run()
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
