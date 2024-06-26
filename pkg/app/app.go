package app

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"github.com/urfave/cli/v2"
	"github.com/zcloud-ws/secure-templates/pkg/config"
	"github.com/zcloud-ws/secure-templates/pkg/connectors"
	"github.com/zcloud-ws/secure-templates/pkg/envs"
	"github.com/zcloud-ws/secure-templates/pkg/helpers"
	"github.com/zcloud-ws/secure-templates/pkg/logging"
	"github.com/zcloud-ws/secure-templates/pkg/render"
	"io"
	"os"
	"time"
)

var (
	appName    = "secure-templates"
	appUsage   = "A template render tool"
	appVersion = "dev"
)

func InitApp(args []string, outfile, errOutfile io.Writer) {
	workdir, err := os.Getwd()
	if err != nil {
		workdir = os.TempDir()
	}
	var cfg config.SecureTemplateConfig
	app := cli.NewApp()
	app.Name = appName
	app.Description = "Secure Templates is a tool to render templates using go-templates and load data values from secrets engine."
	app.Usage = appUsage
	app.Version = appVersion
	app.EnableBashCompletion = true
	var cfgFile, output, secretFile, passphrase string
	var printKeys bool
	if outfile != nil {
		app.Writer = outfile
	}
	if errOutfile != nil {
		app.ErrWriter = errOutfile
	}
	logging.Log.AddHook(&writer.Hook{
		Writer: app.Writer,
		LogLevels: []logrus.Level{
			logrus.DebugLevel,
			logrus.TraceLevel,
			logrus.InfoLevel,
		},
	})
	logging.Log.AddHook(&writer.Hook{
		Writer: app.ErrWriter,
		LogLevels: []logrus.Level{
			logrus.PanicLevel,
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
		Destination: &cfgFile,
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
					Value:       fmt.Sprintf("%s/configs/local-file-secret.json", workdir),
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
				cfg := config.SecureTemplateConfig{
					SecretEngine: config.SecretEngineLocalFile,
					LocalFileConfig: config.LocalFileConfig{
						Filename:   secretFile,
						EncPrivKey: base64.StdEncoding.EncodeToString(privateKey),
						Passphrase: passphrase,
					},
					VaultConfig: config.VaultConfig{
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
						cfg := helpers.ParseConfig(cfgFile, true)
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
						cfg := helpers.ParseConfig(cfgFile, true)
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
		if printKeys {
			cfg = config.SecureTemplateConfig{
				SecretEngine: config.SecretEnginePrintKeys,
			}
		} else {
			cfg = helpers.ParseConfig(cfgFile, false)
		}
		connector = connectors.NewConnector(cfg)
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
			printKeysValues := connector.(*connectors.PrintKeysConnector).Keys
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
