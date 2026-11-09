// Command licenses checks the licenses of the dependencies compiled into the provider.
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli/v3"
	"sigs.k8s.io/yaml"
)

const (
	flagConfigFile     = "config"
	flagIgnorePackages = "ignore-packages"
	flagPlatform       = "platform"
)

const configFile = "internal/cmd/licenses/config.yml"

type config struct {
	Config struct {
		AllowedLicenses []string `json:"allowed-licenses"` // We use json tags so we don't need to import another library as that's how sigs.k8s.io/yaml works.
		IgnoredPackages []string `json:"ignored-packages"`
	} `json:"config"`
}

func main() {
	os.Exit(realMain(os.Args))
}

func realMain(args []string) int {
	app := &cli.Command{
		Name:        "licenses",
		Description: "Checks the licenses of the dependencies compiled into the provider and generates the third-party notices",
		Suggest:     true,
		Commands: []*cli.Command{
			{
				Name:        "check",
				Usage:       "Fail if a dependency uses a license that is not allowed",
				Description: "Checks the licenses of the dependencies compiled into the provider and fails if a dependency uses a license that is not allowed. The allowed licenses are configured in the config file.",
				Flags:       checkFlags,
				Action:      runCheck,
			},
			{
				Name:   "generate",
				Usage:  "Generate the third-party notices",
				Flags:  generateFlags,
				Action: runGenerate,
			},
		},
	}

	if err := app.Run(context.Background(), args); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return 1
	}
	return 0
}

func loadConfig(cmd *cli.Command) (config, error) {
	var cfg config
	path := cmd.String(flagConfigFile)
	b, err := os.ReadFile(path) //nolint:gosec // This is not a security issue; the path is chosen by the person running the tool.
	if err != nil {
		return cfg, fmt.Errorf("reading config file: %w", err)
	}
	if err = yaml.UnmarshalStrict(b, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if cmd.IsSet(flagIgnorePackages) {
		cfg.Config.IgnoredPackages = cmd.StringSlice(flagIgnorePackages)
	}
	return cfg, nil
}

// runGoLicenses runs go-licenses for a GOOS/GOARCH platform.
// The binary is resolved first, as running it through `go tool` would build it for the platform instead of the host.
func runGoLicenses(ctx context.Context, platform string, stdin io.Reader, stdout io.Writer, args ...string) error {
	goos, goarch, ok := strings.Cut(platform, "/")
	if !ok {
		return fmt.Errorf("invalid platform %q, want GOOS/GOARCH", platform)
	}

	var tool bytes.Buffer
	c := exec.CommandContext(ctx, "go", "tool", "-n", "github.com/google/go-licenses/v2")
	c.Stdout = &tool
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("resolving go-licenses: %w", err)
	}

	c = exec.CommandContext(ctx, strings.TrimSpace(tool.String()), args...) //nolint:gosec // The binary was resolved from the go.mod tool dependency.
	c.Env = append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch, "CGO_ENABLED=0")
	c.Stdin = stdin
	c.Stdout = stdout
	c.Stderr = os.Stderr
	return c.Run()
}
