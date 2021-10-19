package main

import (
	//"fmt"
	"log"
	"os"
	"runtime"

	"github.com/urfave/cli/v2"
)

func main() {
	concurrencyN := runtime.NumCPU()

	app := &cli.App{
		Name:  "downloader",
		Usage: "File downloader",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "url",
				Aliases:  []string{"u"},
				Usage:    "`URL` to download",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output `filename`",
			},
			&cli.IntFlag{
				Name:    "concurrency",
				Aliases: []string{"n"},
				Value:   concurrencyN,
				Usage:   "Concurrency `number`",
			},
		},
		Action: func(c *cli.Context) error {
			URL := c.String("url")
			filename := c.String("output")
			concurrency := c.Int("concurrency")
			return NewDownloader(concurrency).Download(URL, filename)
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
