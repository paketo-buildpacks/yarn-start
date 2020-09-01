package main

import (
	"os"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
	yarnstart "github.com/paketo-buildpacks/yarn-start"
)

func main() {
	logger := scribe.NewLogger(os.Stdout)

	packit.Run(
		yarnstart.Detect(),
		yarnstart.Build(logger),
	)
}
