package main

import (
	"log"
	_ "net/http/pprof"
	"os"

	_ "github.com/souvikdeyrit/spinel/pkg/object"
	"github.com/souvikdeyrit/spinel/pkg/utils"
	"github.com/urfave/cli/v2"
)

var logger = utils.GetLogger("spinel")

func main() {
	/*
	   Main driver program to spin up Spinel. It uses a CLI program to read
	   values from command line and mount the volumes.
	*/
	cli.VersionFlag = &cli.BoolFlag{
		Name: "version", Aliases: []string{"V"},
		Usage: "print only the version",
	}
	app := &cli.App{
		Name:    "spinel",
		Usage:   "A POSIX filesystem designed for the cloud native environment.",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"v"},
				Usage:   "enable debug log",
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "only warning and errors",
			},
			&cli.BoolFlag{
				Name:  "trace",
				Usage: "enable trace log",
			},
			&cli.BoolFlag{
				Name:  "nosyslog",
				Usage: "disable syslog",
			},
		},
		Commands: []*cli.Command{
			formatFlags(),
			mountFlags(),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
