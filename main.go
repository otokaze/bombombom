package main

import (
	"os"
	"sort"

	"github.com/urfave/cli/v2"
)

func main() {
	var app = &cli.App{
		Name:                 "bombombom",
		Usage:                "It's not an extremely dangerous toolbox for DDoS and SMS bombing. :)",
		EnableBashCompletion: true,
		Commands: cli.Commands{
			{
				Name:      "ddos",
				Usage:     "ddos attack, and support HTTP proxy mode.",
				ArgsUsage: "<url>",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "requests",
						Usage:   "Number of requests to perform.",
						Aliases: []string{"n"},
						Value:   200,
					},
					&cli.IntFlag{
						Name:    "concurrency",
						Usage:   "Number of multiple requests to make at a time.",
						Aliases: []string{"c"},
						Value:   50,
					},
					&cli.StringFlag{
						Name:    "method",
						Usage:   "HTTP method, one of GET, POST, PUT, DELETE, HEAD, OPTIONS.",
						Aliases: []string{"m"},
						Value:   "GET",
					},
					&cli.StringFlag{
						Name:    "data",
						Usage:   "HTTP request body.",
						Aliases: []string{"d"},
					},
					&cli.StringFlag{
						Name:    "pack",
						Usage:   "Pack of zhima free http proxy.",
						Aliases: []string{"p"},
					},
					&cli.StringSliceFlag{
						Name:    "header",
						Usage:   "Custom HTTP header. For example, -H \"Accept: text/html\" -H \"Content-Type: application/xml\".",
						Aliases: []string{"H"},
					},
				},
				Action: ddosAction,
			},
			{
				Name:  "sms",
				Usage: "sms bombing.",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
		},
		Description: "",
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
