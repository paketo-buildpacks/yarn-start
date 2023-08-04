package main

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	yarnstart "github.com/paketo-buildpacks/yarn-start"
)

func main() {
	logger := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))

	packit.Run(
		yarnstart.Detect(),
		yarnstart.Build(logger),
	)
}
