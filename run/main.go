package main

import (
	"github.com/paketo-buildpacks/packit"
	yarnstart "github.com/paketo-buildpacks/yarn-start"
)

func main() {
	packit.Run(yarnstart.Detect(), yarnstart.Build())
}

