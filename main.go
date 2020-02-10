/*
Drone plugin to upload one or more packages to Bintray.
See README.md for usage.
Author: Archit Sharma December 2019 (Github arcolife)
Previous: David Tootill November 2015 (GitHub tooda02)
*/

package main

import (
	// "log"
	"fmt"
	"os"

	"github.com/arcolife/jfrog-client-go/utils/log"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	Version = "unknown"
	file    *os.File
)

func main() {
	if filename, found := os.LookupEnv("PLUGIN_ENV_FILE"); found {
		_ = godotenv.Load(filename)
	}

	app := cli.NewApp()
	app.Name = "bintray-uploader plugin"
	app.Usage = "bintray-uploader plugin"
	app.Action = run
	app.Version = Version
	app.Flags = []cli.Flag{
		&cli.IntFlag{
			Name:    "bintray.threads,t",
			Usage:   "bintray threads",
			EnvVars: []string{"PLUGIN_BINTRAY_THREADS", "BINTRAY_THREADS"},
			Value:   1,
		},
		&cli.BoolFlag{
			Name:    "bintray.cleanup,D",
			Usage:   "bintray cleanup",
			EnvVars: []string{"PLUGIN_BINTRAY_CLEANUP", "CLEANUP"},
			Value:   false,
		},
		&cli.StringFlag{
			Name:    "bintray.username,u",
			Usage:   "bintray username",
			EnvVars: []string{"PLUGIN_BINTRAY_USERNAME", "USER", "USERNAME", "BINTRAY_USER"},
		},
		&cli.StringFlag{
			Name:    "bintray.config,c",
			Usage:   "bintray config file path",
			EnvVars: []string{"PLUGIN_BINTRAY_CFG", "PACKAGE_CONFIG"},
		},
		&cli.StringFlag{
			Name:    "bintray.api-key",
			Usage:   "bintray api-key",
			EnvVars: []string{"PLUGIN_BINTRAY_API_KEY", "API_KEY", "BINTRAY_KEY"},
		},
		&cli.StringFlag{
			Name:    "bintray.gpg-passphrase,P",
			Usage:   "bintray GPG Passphrase",
			EnvVars: []string{"PLUGIN_BINTRAY_GPG_PASSPHRASE", "GPG_PASSPHRASE", "BINTRAY_ADMIN_GPG_PASSPHRASE"},
		},
		&cli.BoolFlag{
			Name:    "file.upload",
			Usage:   "upload files to package",
			EnvVars: []string{"PLUGIN_BINTRAY_UPLOAD", "UPLOAD_PACKAGE"},
			Value:   true,
		},
		&cli.BoolFlag{
			Name:    "package.sign",
			Usage:   "sign bintray package",
			EnvVars: []string{"PLUGIN_BINTRAY_SIGN", "SIGN_PACKAGE"},
			Value:   true,
		},
		&cli.BoolFlag{
			Name:    "package.publish",
			Usage:   "publish bintray package",
			EnvVars: []string{"PLUGIN_BINTRAY_PUBLISH", "PUBLISH_PACKAGE"},
			Value:   true,
		},
		&cli.BoolFlag{
			Name:    "repo.calc-metadata",
			Usage:   "Calculate Metadata for bintray repo",
			EnvVars: []string{"PLUGIN_BINTRAY_CALC_METADATA", "CALC_META"},
			Value:   true,
		},
		&cli.BoolFlag{
			Name:    "package.show",
			Usage:   "package show",
			EnvVars: []string{"PLUGIN_BINTRAY_SHOW", "SHOW_PACKAGE"},
			Value:   true,
		},
		&cli.StringSliceFlag{
			Name:    "envs",
			Usage:   "pass environment variable to shell script",
			EnvVars: []string{"PLUGIN_ENVS", "INPUT_ENVS"},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	log.SetLogger(log.NewLogger(log.INFO, file))

	fmt.Println("CFG Path:- ", c.String("bintray.config"))

	plugin := Plugin{
		BintrayConfig: BintrayConfig{
			Threads:       c.Int("bintray.threads"),
			Username:      c.String("bintray.username"),
			APIKey:        c.String("bintray.api-key"),
			GPGPassphrase: c.String("bintray.gpg-passphrase"),
		},
		Version:            c.String("version"),
		UploadPackage:      c.Bool("file.upload"),
		Cleanup:            c.Bool("bintray.cleanup"),
		SignPackageVersion: c.Bool("package.sign"),
		PublishPackage:     c.Bool("package.publish"),
		CalcMetadata:       c.Bool("repo.calc-metadata"),
		ShowPackage:        c.Bool("package.show"),
		BintrayConfigPath:  c.String("bintray.config"),
		Envs:               c.StringSlice("envs"),
	}
	return plugin.Exec()
}
