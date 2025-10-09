package main

import (
	"os"
	"strings"

	formationv1 "github.com/gamefabric/gf-core/pkg/api/formation/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/cmd/schema-gen/gen"
	"github.com/hamba/cmd/v3/term"
)

const basePkgPath = "github.com/gamefabric/terraform-provider-gamefabric"

var objs = []gen.PathGroup{
	{
		PrefixPath: "internal/schema/formation",
		ObjInfo: []gen.ObjInfo{
			{
				Filename: "vessel.go",
				Obj:      &formationv1.Vessel{},
			},
			{
				Filename: "formation.go",
				Obj:      &formationv1.Formation{},
			},
		},
	},
}

func main() {
	os.Exit(realMain())
}

func realMain() int {
	ui := newTerm()
	g := gen.New(basePkgPath)

	for _, grp := range objs {
		ui.Output("Generating " + grp.PrefixPath)

		pkgName := prefixPathToPkgName(grp.PrefixPath)
		if err := g.Generate(grp.PrefixPath, pkgName, grp.ObjInfo...); err != nil {
			ui.Error("Error: " + err.Error())
			return 1
		}
	}

	ui.Info("Generation Complete")

	return 0
}

func newTerm() term.Term {
	return term.Prefixed{
		ErrorPrefix: "Error: ",
		Term: term.Colored{
			OutputColor: term.White,
			InfoColor:   term.Cyan,
			ErrorColor:  term.Red,
			Term: term.Basic{
				Writer:      os.Stdout,
				ErrorWriter: os.Stderr,
				Verbose:     true,
			},
		},
	}
}

func prefixPathToPkgName(path string) string {
	lastIdx := strings.LastIndex(path, "/")
	if lastIdx == -1 {
		return path
	}
	return path[lastIdx+1:]
}
