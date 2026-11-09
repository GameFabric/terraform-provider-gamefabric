package main

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"text/template"

	"github.com/ettle/strcase"
	"github.com/urfave/cli/v3"
)

const (
	flagTemplate        = "template"
	flagOut             = "out"
	flagSaveLicensesDir = "save-licenses-dir"
	flagMPLSource       = "include-mpl-source"
)

// licenseFile matches the files that carry license terms or upstream notices.
var licenseFile = regexp.MustCompile(`(?i)^(licen[sc]e|copying|notice|patents)([-._].*)?$`)

var generateFlags = []cli.Flag{
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
		Name:    flagPlatform,
		Usage:   "A GOOS/GOARCH platform to cover. Defaults to the platforms built in .goreleaser.yml",
		Value:   []string{"freebsd/amd64", "freebsd/386", "freebsd/arm", "freebsd/arm64", "windows/amd64", "windows/386", "windows/arm64", "linux/amd64", "linux/386", "linux/arm", "linux/arm64", "darwin/amd64", "darwin/arm64"},
		Sources: cli.EnvVars(strcase.ToSNAKE(flagPlatform)),
	},
	&cli.StringFlag{
		Name:    flagTemplate,
		Usage:   "Path to the notices template",
		Value:   "internal/cmd/licenses/THIRD_PARTY_NOTICES.tmpl.md",
		Sources: cli.EnvVars(strcase.ToSNAKE(flagTemplate)),
	},
	&cli.StringFlag{
		Name:    flagOut,
		Usage:   "Path of the notices file to write",
		Value:   "THIRD_PARTY_NOTICES.md",
		Sources: cli.EnvVars(strcase.ToSNAKE(flagOut)),
	},
	&cli.StringFlag{
		Name:    flagSaveLicensesDir,
		Usage:   "Directory to save the license texts and upstream notices into, such as LICENSES/. If empty, they are not saved",
		Sources: cli.EnvVars(strcase.ToSNAKE(flagSaveLicensesDir)),
	},
	&cli.BoolFlag{
		Name:    flagMPLSource,
		Usage:   "Keep the full source of MPL-2.0 dependencies in the licenses directory instead of only their license files",
		Sources: cli.EnvVars(strcase.ToSNAKE(flagMPLSource)),
	},
}

func runGenerate(ctx context.Context, cmd *cli.Command) error {
	cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}

	var ignore []string
	for _, pkg := range cfg.Config.IgnoredPackages {
		ignore = append(ignore, "--ignore="+pkg)
	}

	// The Go version comes from go.mod, as that is what the release is built with, regardless of the local toolchain.
	var goInfo [2]string
	for i, args := range [][]string{{"list", "-m", "-f", "go{{.GoVersion}}"}, {"env", "GOROOT"}} {
		var out bytes.Buffer
		c := exec.CommandContext(ctx, "go", args...) //nolint:gosec // The arguments are fixed values.
		c.Stdout = &out
		c.Stderr = os.Stderr
		if err = c.Run(); err != nil {
			return fmt.Errorf("resolving go environment: %w", err)
		}
		goInfo[i] = strings.TrimSpace(out.String())
	}
	goVersion, goRoot := goInfo[0], goInfo[1]

	// The template is rendered twice. First here, to fill in the Go version and to get the row template go-licenses renders, and then to fill in the merged rows.
	path := cmd.String(flagTemplate)
	tmpl, err := template.New(filepath.Base(path)).Option("missingkey=error").ParseFiles(path)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	var rowTmpl bytes.Buffer
	if err = tmpl.ExecuteTemplate(&rowTmpl, "ROWS", nil); err != nil {
		return fmt.Errorf("rendering row template: %w", err)
	}

	var rows []string
	for _, platform := range cmd.StringSlice(flagPlatform) {
		var report bytes.Buffer
		args := append([]string{"report", ".", "--template=/dev/stdin"}, ignore...)
		if err = runGoLicenses(ctx, platform, bytes.NewReader(rowTmpl.Bytes()), &report, args...); err != nil {
			return fmt.Errorf("reporting licenses for %s: %w", platform, err)
		}
		for row := range strings.SplitSeq(report.String(), "\n") {
			if strings.TrimSpace(row) != "" {
				rows = append(rows, row)
			}
		}
	}
	slices.Sort(rows)

	var out bytes.Buffer
	data := map[string]string{"GO_VERSION": goVersion, "GO_LICENSE_REPORT": strings.Join(slices.Compact(rows), "\n")}
	if err = tmpl.Execute(&out, data); err != nil {
		return fmt.Errorf("rendering template: %w", err)
	}
	if err = os.WriteFile(cmd.String(flagOut), out.Bytes(), 0o644); err != nil { //nolint:gosec // The notices are meant to be readable by everyone.
		return err
	}

	if cmd.String(flagSaveLicensesDir) == "" {
		return nil
	}
	return saveLicenses(ctx, cmd, ignore, goRoot)
}

func saveLicenses(ctx context.Context, cmd *cli.Command, ignore []string, goRoot string) error {
	dir := cmd.String(flagSaveLicensesDir)
	tmp, err := os.MkdirTemp("", "licenses-")
	if err != nil {
		return err
	}
	// The saved directories keep the read-only permissions of the module cache.
	defer func() {
		_ = filepath.WalkDir(tmp, func(path string, d fs.DirEntry, err error) error {
			if err == nil && d.IsDir() {
				return os.Chmod(path, 0o700) //nolint:gosec // Directories need to be traversable to be removed.
			}
			return nil
		})
		_ = os.RemoveAll(tmp)
	}()

	// go-licenses can only save into a directory that doesn't exist, so each platform is saved on its own and merged.
	if err = os.RemoveAll(dir); err != nil {
		return err
	}
	if err = os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	for _, platform := range cmd.StringSlice(flagPlatform) {
		src := filepath.Join(tmp, strings.ReplaceAll(platform, "/", "_"))
		if err = runGoLicenses(ctx, platform, nil, nil, append([]string{"save", ".", "--save_path=" + src}, ignore...)...); err != nil {
			return fmt.Errorf("saving licenses for %s: %w", platform, err)
		}
		if err = mergeDir(src, dir); err != nil {
			return err
		}
	}

	if !cmd.Bool(flagMPLSource) {
		if err = keepLicenseFiles(dir); err != nil {
			return err
		}
	}
	return copyGoLicenses(goRoot, dir)
}

// mergeDir copies the files of src into dst, keeping the files that dst already has.
func mergeDir(src, dst string) error {
	srcRoot, err := os.OpenRoot(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcRoot.Close() }()

	dstRoot, err := os.OpenRoot(dst)
	if err != nil {
		return err
	}
	defer func() { _ = dstRoot.Close() }()

	return fs.WalkDir(srcRoot.FS(), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return dstRoot.MkdirAll(path, 0o750)
		}
		if _, err = dstRoot.Stat(path); err == nil {
			return nil
		}
		b, err := srcRoot.ReadFile(path)
		if err != nil {
			return err
		}
		return dstRoot.WriteFile(path, b, 0o644)
	})
}

// copyGoLicenses copies the license and patent grant of the Go runtime into dir.
func copyGoLicenses(goRoot, dir string) error {
	if err := os.MkdirAll(filepath.Join(dir, "go"), 0o750); err != nil {
		return err
	}

	for _, name := range []string{"LICENSE", "PATENTS"} {
		// Some distributions, such as Homebrew, keep these outside of GOROOT.
		src := filepath.Join(goRoot, name)
		if _, err := os.Stat(src); err != nil {
			src = filepath.Join(goRoot, "..", name)
		}
		b, err := os.ReadFile(src) //nolint:gosec // This is not a security issue; the file is part of the local Go installation.
		if err != nil {
			return fmt.Errorf("reading go %s: %w", name, err)
		}
		if err = os.WriteFile(filepath.Join(dir, "go", name), b, 0o644); err != nil { //nolint:gosec // The license files are meant to be readable by everyone.
			return err
		}
	}
	return nil
}

// keepLicenseFiles removes everything but license and notice files from dir, which drops the sources go-licenses copies for MPL-2.0 dependencies.
func keepLicenseFiles(dir string) error {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()

	var dirs []string
	err = fs.WalkDir(root.FS(), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			dirs = append(dirs, path)
			return nil
		}
		if licenseFile.MatchString(d.Name()) {
			return nil
		}
		return root.Remove(path)
	})
	if err != nil {
		return err
	}

	// Remove the directories left empty, deepest first. Non-empty ones are kept.
	slices.Reverse(dirs)
	for _, d := range dirs {
		_ = root.Remove(d)
	}
	return nil
}
