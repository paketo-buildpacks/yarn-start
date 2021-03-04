package main

import (
	"os"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
	yarnstart "github.com/paketo-buildpacks/yarn-start"
)

func main() {
	logger := scribe.NewLogger(os.Stdout)
	projectPathParser := yarnstart.NewProjectPathParser()

	packit.Run(
		yarnstart.Detect(projectPathParser),
		yarnstart.Build(projectPathParser, logger),
	)
}
