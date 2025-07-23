package main

import (
	"log"
	"os"

	"github.com/artificialinc/docker-context-size/pkg/docker"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "docker-context-size",
		Usage: "Display Docker build context contents in a tree format",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "directory",
				Aliases: []string{"d"},
				Value:   ".",
				Usage:   "Directory to analyze",
			},
			&cli.IntFlag{
				Name:    "depth",
				Aliases: []string{"l"},
				Value:   1,
				Usage:   "Maximum depth to display (-1 for unlimited)",
			},
		},
		Action: func(c *cli.Context) error {
			dir := c.String("directory")
			depth := c.Int("depth")

			return docker.BuildLocalContext(dir, depth)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
