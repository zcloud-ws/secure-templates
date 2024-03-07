package main

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	config2 "github.com/edimarlnx/secure-templates/pkg/config"
	"github.com/edimarlnx/secure-templates/pkg/connectors"
	"github.com/edimarlnx/secure-templates/pkg/helpers"
	"github.com/edimarlnx/secure-templates/pkg/render"
	"github.com/urfave/cli/v2"
	"io"
	"log/slog"
	"os"
	"time"
)

var (
	appName    = "secure-templates"
	appUsage   = "A template render tool"
	appVersion = "dev"
)

func main() {
	initApp(os.Args, nil)
}

func initApp(args []string, outfile io.Writer) {
	workdir, err := os.Getwd()
	if err != nil {
		workdir = os.TempDir()
	}
	app := cli.NewApp()
	app.Name = appName
	app.Description = "Secure Templates is a tool to render templates using go-templates and load data values from secrets engine."
	app.Usage = appUsage
	app.Version = appVersion
	app.EnableBashCompletion = true
	var config, output, secretFile, passphrase string
	if outfile != nil {
		app.Writer = outfile
	}
	configFlag := cli.StringFlag{
		Name:        "config",
		Aliases:     []string{"c", "cfg"},
		EnvVars:     []string{"SEC_TPL_CONFIG"},
		Value:       "",
		Destination: &config,
	}
	outputFlag := cli.StringFlag{
		Name:        "output",
		Aliases:     []string{"o", "out"},
		EnvVars:     []string{"SEC_TPL_OUTPUT"},
		Value:       "",
		Destination: &output,
	}

	app.Commands = []*cli.Command{
		{
			Name:  "init-config",
			Usage: "Init a sample config",
			Flags: []cli.Flag{
				&outputFlag,
				&cli.StringFlag{
					Name:        "secret-file",
					Value:       fmt.Sprintf("%s/local-file-secret.json", workdir),
					Destination: &secretFile,
				},
				&cli.StringFlag{
					Name:        "private-key-passphrase",
					EnvVars:     []string{"LOCAL_SECRET_PRIVATE_KEY_PASSPHRASE"},
					Value:       "",
					Destination: &passphrase,
				},
			},
			Action: func(cCtx *cli.Context) error {
				if passphrase == "" {
					passphrase = fmt.Sprintf("%x", md5.Sum([]byte(time.Now().String())))
				}
				privateKey, err := helpers.GenRsaPrivateKey(passphrase)
				cfg := config2.SecureTemplateConfig{
					SecretEngine: config2.SecretEngineLocalFile,
					LocalFileConfig: config2.LocalFileConfig{
						Filename:   secretFile,
						EncPrivKey: base64.StdEncoding.EncodeToString(privateKey),
						Passphrase: passphrase,
					},
					VaultConfig: config2.VaultConfig{
						Address:      "http://localhost:8200",
						Token:        "token",
						SecretEngine: "kv",
						Namespace:    "dev",
					},
				}
				outJson := cCtx.App.Writer
				if output != "" {
					file, err := os.Create(output)
					if err != nil {
						return err
					}
					outJson = file
				}
				err = cfg.Json(outJson)
				return err
			},
		},
		{
			Name:  "local-secret",
			Usage: "Manage local secret file",
			Flags: []cli.Flag{
				&configFlag,
			},
			Subcommands: []*cli.Command{
				{
					Name:      "put",
					Usage:     "Add or update key value",
					UsageText: "put SECRET KEY VALUE",
					ArgsUsage: "[secret and key and value]",
					Args:      true,
					Action: func(cCtx *cli.Context) error {
						if len(cCtx.Args().Slice()) < 3 {
							return cli.Exit("Required secret, key and value args", 1)
						}
						cfg := helpers.ParseConfig(config)
						if cfg.SecretEngine != config2.SecretEngineLocalFile {
							return cli.Exit("local-secret command requires local-file as secret engine.", 1)
						}
						connector := connectors.NewConnector(cfg)
						secret := cCtx.Args().Get(0)
						key := cCtx.Args().Get(1)
						value := cCtx.Args().Get(2)
						err := connector.WriteKey(secret, key, value)
						if err == nil {
							cCtx.App.Writer.Write([]byte(fmt.Sprintf("Key '%s' saved on secret '%s'\n", key, secret)))
						}
						return err
					},
				},
			},
		},
	}
	app.Flags = []cli.Flag{
		&configFlag,
		&outputFlag,
	}
	app.Action = func(c *cli.Context) error {
		if _, err := os.Stat(config); os.IsNotExist(err) {
			return cli.Exit(fmt.Sprintf("Config file not found: %s", config), 1)
		}
		cfg := helpers.ParseConfig(config)
		connector := connectors.NewConnector(cfg)
		filename := c.Args().First()
		file, err := os.Open(filename)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Error on open input file %s", filename), 1)
		}
		outputFile := c.App.Writer
		if output != "" && output != "-" {
			outputFile, err = os.Create(output)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Error on open output file %s", filename), 1)
			}
		}
		err = render.ParseFile(file, connector, outputFile)
		if err != nil {
			return cli.Exit(err.Error(), 1)
		}

		return nil
	}
	appArgs := args
	if len(args) < 2 {
		appArgs = append(args, "-h")
	}
	if err := app.Run(appArgs); err != nil {
		slog.Error(err.Error())
	}
}
