package main

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	config2 "github.com/edimarlnx/secure-templates/pkg/config"
	"github.com/edimarlnx/secure-templates/pkg/connectors"
	"github.com/edimarlnx/secure-templates/pkg/envs"
	"github.com/edimarlnx/secure-templates/pkg/helpers"
	"github.com/edimarlnx/secure-templates/pkg/logging"
	"github.com/edimarlnx/secure-templates/pkg/render"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"github.com/urfave/cli/v2"
	"io"
	"os"
	"time"
)

var (
	appName    = "secure-templates"
	appUsage   = "A template render tool"
	appVersion = "dev"
)

func main() {
	initApp(os.Args, nil, nil)
}

func initApp(args []string, outfile, errOutfile io.Writer) {
	workdir, err := os.Getwd()
	if err != nil {
		workdir = os.TempDir()
	}
	var cfg config2.SecureTemplateConfig
	app := cli.NewApp()
	app.Name = appName
	app.Description = "Secure Templates is a tool to render templates using go-templates and load data values from secrets engine."
	app.Usage = appUsage
	app.Version = appVersion
	app.EnableBashCompletion = true
	var config, output, secretFile, passphrase string
	var printKeys bool
	if outfile != nil {
		app.Writer = outfile
	}
	if errOutfile != nil {
		app.ErrWriter = errOutfile
	}
	logging.Log.SetOutput(app.Writer)
	logging.Log.AddHook(&writer.Hook{
		Writer: app.ErrWriter,
		LogLevels: []logrus.Level{
			logrus.WarnLevel,
			logrus.ErrorLevel,
			logrus.FatalLevel,
		},
	})
	configFlag := cli.StringFlag{
		Name:        "config",
		Aliases:     []string{"c", "cfg"},
		EnvVars:     []string{envs.SecTplConfigEnv},
		Value:       "",
		Destination: &config,
	}
	outputFlag := cli.StringFlag{
		Name:        "output",
		Aliases:     []string{"o", "out"},
		EnvVars:     []string{envs.SecTplOutputEnv},
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
					EnvVars:     []string{envs.LocalSecretPrivateKeyPassphraseEnv},
					Value:       "",
					Destination: &passphrase,
				},
			},
			Action: func(cCtx *cli.Context) error {
				if passphrase == "" {
					passphrase = fmt.Sprintf("%x", md5.Sum([]byte(time.Now().String())))
				}
				var privateKey []byte
				privateKey, err = helpers.GenRsaPrivateKey(passphrase)
				if err != nil {
					return err
				}
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
			Name:  "manage-secret",
			Usage: "Manage secret",
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
						connector := connectors.NewConnector(cfg)
						secret := cCtx.Args().Get(0)
						key := cCtx.Args().Get(1)
						value := cCtx.Args().Get(2)
						err := connector.WriteKey(secret, key, value)
						if err == nil {
							logging.Log.Infof("Key '%s' saved on secret '%s'\n", key, secret)
						}
						return err
					},
				},
				{
					Name:      "import",
					Usage:     "Add or update key value using env file",
					UsageText: "import filepath",
					ArgsUsage: "[import and filepath]",
					Args:      true,
					Action: func(cCtx *cli.Context) error {
						if len(cCtx.Args().Slice()) < 2 {
							return cli.Exit("Required filename and secret args", 1)
						}
						envFile := cCtx.Args().Get(1)
						data, err := helpers.ParseEnvFileAsKeyValue(envFile)
						if err != nil {
							return cli.Exit(err.Error(), 1)
						}
						cfg := helpers.ParseConfig(config)
						connector := connectors.NewConnector(cfg)
						secret := cCtx.Args().Get(0)
						err = connector.WriteKeys(secret, data)
						if err == nil {
							logging.Log.Infof("%d keys saved on secret '%s'\n", len(data), secret)
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
		&cli.BoolFlag{
			Name:        "print-keys",
			Aliases:     []string{"p"},
			Value:       false,
			Destination: &printKeys,
		},
	}
	app.Action = func(c *cli.Context) error {
		var connector connectors.Connector
		var printKeysValues map[string]int
		if printKeys {
			printKeysValues = map[string]int{}
			connector = &connectors.PrintKeysConnector{
				Keys: printKeysValues,
			}
		} else {
			if _, err := os.Stat(config); os.IsNotExist(err) {
				return cli.Exit(fmt.Sprintf("Config file not found: %s", config), 1)
			}
			cfg = helpers.ParseConfig(config)
			connector = connectors.NewConnector(cfg)
		}
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
		//logging.Log.AddHook(&writer.Hook{
		//	Writer: c.App.ErrWriter,
		//	LogLevels: []logrus.Level{
		//		logrus.WarnLevel,
		//	},
		//})
		if printKeys {
			nullOutput, err := os.Create(os.DevNull)
			if err != nil {
				return cli.Exit("Error on open output file /dev/null", 1)
			}
			err = render.ParseFile(cfg.Options, file, connector, nullOutput)
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}
			_, err = outputFile.Write([]byte("Template keys:\n"))
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}
			for key := range printKeysValues {
				_, err := outputFile.Write([]byte("  " + key + "\n"))
				if err != nil {
					return cli.Exit(err.Error(), 1)
				}
			}

		} else {
			err := render.ParseFile(cfg.Options, file, connector, outputFile)
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}
		}
		return nil
	}
	appArgs := args
	if len(args) < 2 {
		appArgs = append(args, "-h")
	}
	if err := app.Run(appArgs); err != nil {
		logging.Log.Error(err.Error())
	}
}
