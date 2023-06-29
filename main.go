package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"c-docker/config"
)

const usage = `这是一个容器引擎`

func main() {
	config.GlobalExeLocation, _ = os.Getwd()
	app := cli.NewApp()
	app.Name = "c-docker"
	app.Usage = usage

	app.Commands = []cli.Command{
		initCommand,
		runCommand,
		listCommand,
		logCommand,
		execCommand,
		stopCommand,
		removeCommand,
		commitCommand,
		networkCommand,
	}

	app.Before = func(context *cli.Context) error {
		// Log as JSON instead of the default ASCII formatter.
		log.SetFormatter(&log.JSONFormatter{})

		log.SetOutput(os.Stdout)
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
