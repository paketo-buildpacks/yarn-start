package yarnstart

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"
)

//go:generate faux --interface PathParser --output fakes/path_parser.go
type PathParser interface {
	Get(path string) (projectPath string, err error)
}

// NoStartScriptError indicates that the targeted project does no have a start command in their package.json
const NoStartScriptError = "no start script in package.json"

func Detect(projectPathParser PathParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		projectPath, err := projectPathParser.Get(context.WorkingDir)
		if err != nil {
			return packit.DetectResult{}, err
		}

		exists, err := fs.Exists(filepath.Join(projectPath, "yarn.lock"))
		if err != nil {
			return packit.DetectResult{}, fmt.Errorf("failed to stat yarn.lock: %w", err)
		}

		if !exists {
			return packit.DetectResult{}, packit.Fail.WithMessage(`no "yarn.lock" found in the project path %s`, projectPath)
		}

		var pkg *PackageJson
		if pkg, err = NewPackageJsonFromPath(filepath.Join(projectPath, "package.json")); err != nil {
			return packit.DetectResult{}, err
		}

		if !pkg.hasStartCommand() {
			return packit.DetectResult{}, packit.Fail.WithMessage(NoStartScriptError)
		}

		requirements := []packit.BuildPlanRequirement{
			{
				Name: Node,
				Metadata: map[string]interface{}{
					"launch": true,
				},
			},
			{
				Name: Yarn,
				Metadata: map[string]interface{}{
					"launch": true,
				},
			},
			{
				Name: NodeModules,
				Metadata: map[string]interface{}{
					"launch": true,
				},
			},
		}

		shouldReload, err := checkLiveReloadEnabled()
		if err != nil {
			return packit.DetectResult{}, err
		}

		if shouldReload {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "watchexec",
				Metadata: map[string]interface{}{
					"launch": true,
				},
			})
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Requires: requirements,
			},
		}, nil
	}
}

func checkLiveReloadEnabled() (bool, error) {
	if reload, ok := os.LookupEnv("BP_LIVE_RELOAD_ENABLED"); ok {
		shouldEnableReload, err := strconv.ParseBool(reload)
		if err != nil {
			return false, fmt.Errorf("failed to parse BP_LIVE_RELOAD_ENABLED value %s: %w", reload, err)
		}
		return shouldEnableReload, nil
	}
	return false, nil
}
