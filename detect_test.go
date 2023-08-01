package yarnstart_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	yarnstart "github.com/paketo-buildpacks/yarn-start"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string
		detect     packit.DetectFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Mkdir(filepath.Join(workingDir, "custom"), os.ModePerm)).To(Succeed())

		t.Setenv("BP_NODE_PROJECT_PATH", "custom")

		detect = yarnstart.Detect()
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("when there is a yarn.lock and a start script", func() {
		it.Before(func() {
			Expect(os.WriteFile(filepath.Join(workingDir, "custom", "package.json"), []byte(`{
				"scripts": {
					"start": "node server.js"
				}
			}`), 0600)).To(Succeed())

			Expect(os.WriteFile(filepath.Join(workingDir, "custom", "yarn.lock"), nil, 0600)).To(Succeed())
		})

		it("detects", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"launch": true,
						},
					},
					{
						Name: "yarn",
						Metadata: map[string]interface{}{
							"launch": true,
						},
					},
					{
						Name: "node_modules",
						Metadata: map[string]interface{}{
							"launch": true,
						},
					},
				},
			}))
		})

		context("and BP_LIVE_RELOAD_ENABLED=true in the build environment", func() {
			it.Before(func() {
				os.Setenv("BP_LIVE_RELOAD_ENABLED", "true")
			})

			it.After(func() {
				os.Unsetenv("BP_LIVE_RELOAD_ENABLED")
			})

			it("requires watchexec at launch time", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan.Requires).To(Equal([]packit.BuildPlanRequirement{
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"launch": true,
						},
					},
					{
						Name: "yarn",
						Metadata: map[string]interface{}{
							"launch": true,
						},
					},
					{
						Name: "node_modules",
						Metadata: map[string]interface{}{
							"launch": true,
						},
					},
					{
						Name: "watchexec",
						Metadata: map[string]interface{}{
							"launch": true,
						},
					},
				},
				))
			})
		})
	})

	context("when there is no yarn.lock", func() {
		it("fails detection", func() {
			_, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).To(MatchError(packit.Fail.WithMessage("no 'yarn.lock' found in the project path %s", filepath.Join(workingDir, "custom"))))
		})
	})

	context("failure cases", func() {
		context("the workspace directory cannot be accessed", func() {
			it.Before(func() {
				Expect(os.Chmod(workingDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the project path cannot be found", func() {
			it.Before(func() {
				t.Setenv("BP_NODE_PROJECT_PATH", "does-not-exist")
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(ContainSubstring("could not find project path")))
			})
		})

		context("when BP_LIVE_RELOAD_ENABLED is set to an invalid value", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(workingDir, "custom", "package.json"), []byte(`{
					"scripts": {
						"start": "node server.js"
					}
				}`), 0600)).To(Succeed())

				Expect(os.WriteFile(filepath.Join(workingDir, "custom", "yarn.lock"), nil, 0600)).To(Succeed())
				os.Setenv("BP_LIVE_RELOAD_ENABLED", "not-a-bool")
			})

			it.After(func() {
				os.Unsetenv("BP_LIVE_RELOAD_ENABLED")
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(ContainSubstring("failed to parse BP_LIVE_RELOAD_ENABLED value not-a-bool")))
			})
		})
	})
}
