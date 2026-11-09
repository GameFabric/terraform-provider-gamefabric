package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ettle/strcase"
	"github.com/urfave/cli/v3"
)

const flagAllowedLicenses = "allowed-licenses"

var checkFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    flagConfigFile,
		Usage:   "Path to the configuration file",
		Value:   configFile,
		Sources: cli.EnvVars(strcase.ToSNAKE(flagConfigFile)),
	},
	&cli.StringSliceFlag{
		Name:    flagIgnorePackages,
		Usage:   "A package path prefix to leave out of the check, such as our own packages",
		Sources: cli.EnvVars(strcase.ToSNAKE(flagIgnorePackages)),
	},
	&cli.StringSliceFlag{
		Name:  flagPlatform,
		Usage: "A GOOS/GOARCH platform to cover. Defaults to the platforms built in .goreleaser.yml",
		Value: []string{
			"freebsd/amd64", "freebsd/386", "freebsd/arm", "freebsd/arm64",
			"windows/amd64", "windows/386", "windows/arm64",
			"linux/amd64", "linux/386", "linux/arm", "linux/arm64",
			"darwin/amd64", "darwin/arm64",
		},
		Sources: cli.EnvVars(strcase.ToSNAKE(flagPlatform)),
	},
	&cli.StringSliceFlag{
		Name:    flagAllowedLicenses,
		Usage:   "A license dependencies are allowed to use",
		Sources: cli.EnvVars(strcase.ToSNAKE(flagAllowedLicenses)),
	},
}

func runCheck(ctx context.Context, cmd *cli.Command) error {
	cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}
	if cmd.IsSet(flagAllowedLicenses) {
		cfg.Config.AllowedLicenses = cmd.StringSlice(flagAllowedLicenses)
	}

	args := []string{"check", "."}
	if len(cfg.Config.AllowedLicenses) > 0 {
		args = append(args, "--allowed_licenses="+strings.Join(cfg.Config.AllowedLicenses, ","))
	}
	for _, pkg := range cfg.Config.IgnoredPackages {
		args = append(args, "--ignore="+pkg)
	}

	for _, platform := range cmd.StringSlice(flagPlatform) {
		if err = runGoLicenses(ctx, platform, nil, os.Stdout, args...); err != nil {
			return fmt.Errorf("checking licenses for %s: %w", platform, err)
		}
	}
	return nil
}
