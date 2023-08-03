package yarnstart_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	yarnstart "github.com/paketo-buildpacks/yarn-start"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir  string
		workingDir string
		cnbDir     string
		buffer     *bytes.Buffer

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = os.MkdirTemp("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.Mkdir(filepath.Join(workingDir, "some-project-dir"), os.ModePerm)).To(Succeed())
		err = os.WriteFile(filepath.Join(workingDir, "some-project-dir", "package.json"), []byte(`{
			"scripts": {
				"prestart": "some-prestart-command",
				"start": "some-start-command",
				"poststart": "some-poststart-command"
			}
		}`), 0600)
		Expect(err).NotTo(HaveOccurred())

		buffer = bytes.NewBuffer(nil)
		logger := scribe.NewEmitter(buffer)

		build = yarnstart.Build(logger)
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("returns a result that builds correctly", func() {
		it.Before(func() {
			t.Setenv("BP_NODE_PROJECT_PATH", "some-project-dir")
		})

		it("default", func() {
			result, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(packit.BuildResult{
				Launch: packit.LaunchMetadata{
					Processes: []packit.Process{
						{
							Type:    "web",
							Command: "bash",
							Args: []string{
								"-c",
								fmt.Sprintf("cd %s/some-project-dir && some-prestart-command && some-start-command && some-poststart-command", workingDir),
							},
							Default: true,
							Direct:  true,
						},
					},
				},
			}))
		})
	})

	context("when BP_LIVE_RELOAD_ENABLED=true in the build environment", func() {
		it.Before(func() {
			t.Setenv("BP_LIVE_RELOAD_ENABLED", "true")
			t.Setenv("BP_NODE_PROJECT_PATH", "some-project-dir")
		})

		it("adds a reloadable start command that ignores package manager files and makes it the default", func() {
			result, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Launch.Processes).To(Equal([]packit.Process{
				{
					Type:    "web",
					Command: "watchexec",
					Args: []string{
						"--restart",
						"--shell", "none",
						"--watch", filepath.Join(workingDir, "some-project-dir"),
						"--ignore", filepath.Join(workingDir, "some-project-dir", "package.json"),
						"--ignore", filepath.Join(workingDir, "some-project-dir", "yarn.lock"),
						"--ignore", filepath.Join(workingDir, "some-project-dir", "node_modules"),
						"--",
						"bash", "-c",
						fmt.Sprintf("cd %s/some-project-dir && some-prestart-command && some-start-command && some-poststart-command", workingDir),
					},
					Default: true,
					Direct:  true,
				},
				{
					Type:    "no-reload",
					Command: "bash",
					Args: []string{
						"-c",
						fmt.Sprintf("cd %s/some-project-dir && some-prestart-command && some-start-command && some-poststart-command", workingDir),
					},
					Direct: true,
				},
			}))
		})
	})

	context("when the package.json does not include a prestart command", func() {
		it.Before(func() {
			err := os.WriteFile(filepath.Join(workingDir, "some-project-dir", "package.json"), []byte(`{
				"scripts": {
					"start": "some-start-command",
					"poststart": "some-poststart-command"
				}
			}`), 0600)
			Expect(err).NotTo(HaveOccurred())
			t.Setenv("BP_NODE_PROJECT_PATH", "some-project-dir")
		})

		it("returns a result with a valid start command", func() {
			result, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(packit.BuildResult{
				Launch: packit.LaunchMetadata{
					Processes: []packit.Process{
						{
							Type:    "web",
							Command: "bash",
							Args: []string{
								"-c",
								fmt.Sprintf("cd %s/some-project-dir && some-start-command && some-poststart-command", workingDir),
							},
							Default: true,
							Direct:  true,
						},
					}},
			}))
		})
	})

	context("when the package.json does not include a poststart command", func() {
		it.Before(func() {
			err := os.WriteFile(filepath.Join(workingDir, "some-project-dir", "package.json"), []byte(`{
				"scripts": {
					"prestart": "some-prestart-command",
					"start": "some-start-command"
				}
			}`), 0600)
			Expect(err).NotTo(HaveOccurred())
			t.Setenv("BP_NODE_PROJECT_PATH", "some-project-dir")
		})

		it("returns a result with a valid start command", func() {
			result, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(packit.BuildResult{
				Launch: packit.LaunchMetadata{
					Processes: []packit.Process{
						{
							Type:    "web",
							Command: "bash",
							Args: []string{
								"-c",
								fmt.Sprintf("cd %s/some-project-dir && some-prestart-command && some-start-command", workingDir),
							},
							Default: true,
							Direct:  true,
						},
					},
				},
			}))
		})
	})

	context("when the package.json does not include a start command", func() {
		it.Before(func() {
			err := os.WriteFile(filepath.Join(workingDir, "some-project-dir", "package.json"), []byte(`{
				"scripts": {
					"prestart": "some-prestart-command",
					"poststart": "some-poststart-command"
				}
			}`), 0600)
			Expect(err).NotTo(HaveOccurred())
			t.Setenv("BP_NODE_PROJECT_PATH", "some-project-dir")
		})

		it("returns a result with a valid start command", func() {
			result, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(packit.BuildResult{
				Launch: packit.LaunchMetadata{
					Processes: []packit.Process{
						{
							Type:    "web",
							Command: "bash",
							Args: []string{
								"-c",
								fmt.Sprintf("cd %[1]s/some-project-dir && some-prestart-command && node %[1]s/server.js && some-poststart-command", workingDir),
							},
							Default: true,
							Direct:  true,
						},
					},
				},
			}))
		})
	})

	context("when the project-path env var is not set", func() {
		it.Before(func() {
			err := os.WriteFile(filepath.Join(workingDir, "package.json"), []byte(`{
				"scripts": {
					"prestart": "some-prestart-command",
					"start": "some-start-command",
					"poststart": "some-poststart-command"
				}
			}`), 0600)
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(os.Remove(filepath.Join(workingDir, "package.json"))).To(Succeed())
		})

		it("returns a result with a valid start command", func() {
			result, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(packit.BuildResult{
				Launch: packit.LaunchMetadata{
					Processes: []packit.Process{
						{
							Type:    "web",
							Command: "bash",
							Args: []string{
								"-c",
								"some-prestart-command && some-start-command && some-poststart-command",
							},
							Default: true,
							Direct:  true,
						},
					},
				},
			}))
		})
	})

	context("failure cases", func() {
		context("when package.json does not exist", func() {
			it.Before(func() {
				Expect(os.RemoveAll(filepath.Join(workingDir, "some-project-dir", "package.json"))).To(Succeed())
				t.Setenv("BP_NODE_PROJECT_PATH", "some-project-dir")
			})

			it("fails with the appropriate error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					CNBPath:    cnbDir,
					Stack:      "some-stack",
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "some-version",
					},
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("open %s: no such file or directory", filepath.Join(workingDir, "some-project-dir", "package.json")))))
			})
		})

		context("when package.json is malformed", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(workingDir, "some-project-dir", "package.json"), []byte("%%%"), os.ModePerm)).To(Succeed())
				t.Setenv("BP_NODE_PROJECT_PATH", "some-project-dir")
			})

			it("fails with the appropriate error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					CNBPath:    cnbDir,
					Stack:      "some-stack",
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "some-version",
					},
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError(ContainSubstring("unable to decode package.json invalid character '%' looking for beginning of value")))
			})
		})

		context("when BP_LIVE_RELOAD_ENABLED is set to an invalid value", func() {
			it.Before(func() {
				t.Setenv("BP_LIVE_RELOAD_ENABLED", "not-a-bool")
				t.Setenv("BP_NODE_PROJECT_PATH", "some-project-dir")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					CNBPath:    cnbDir,
					Stack:      "some-stack",
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "some-version",
					},
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError(ContainSubstring("failed to parse BP_LIVE_RELOAD_ENABLED value not-a-bool")))
			})
		})
	})
}
