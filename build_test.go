package yarnstart_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
	yarnstart "github.com/paketo-buildpacks/yarn-start"
	"github.com/paketo-buildpacks/yarn-start/fakes"
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
		pathParser *fakes.PathParser

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = ioutil.TempDir("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = ioutil.TempDir("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.Mkdir(filepath.Join(workingDir, "some-project-dir"), os.ModePerm)).To(Succeed())
		err = ioutil.WriteFile(filepath.Join(workingDir, "some-project-dir", "package.json"), []byte(`{
			"scripts": {
				"prestart": "some-prestart-command",
				"start": "some-start-command",
				"poststart": "some-poststart-command"
			}
		}`), 0644)
		Expect(err).NotTo(HaveOccurred())

		buffer = bytes.NewBuffer(nil)
		logger := scribe.NewLogger(buffer)

		pathParser = &fakes.PathParser{}

		pathParser.GetCall.Returns.ProjectPath = filepath.Join(workingDir, "some-project-dir")
		build = yarnstart.Build(pathParser, logger)
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("returns a result that builds correctly", func() {
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
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{},
			},
			Launch: packit.LaunchMetadata{
				Processes: []packit.Process{
					{
						Type:    "web",
						Command: fmt.Sprintf("cd %s && some-prestart-command && some-start-command && some-poststart-command", filepath.Join(workingDir, "some-project-dir")),
					},
				},
			},
		}))
		Expect(pathParser.GetCall.Receives.Path).To(Equal(workingDir))
	})

	context("when the package.json does not include a prestart command", func() {
		it.Before(func() {
			err := ioutil.WriteFile(filepath.Join(workingDir, "some-project-dir", "package.json"), []byte(`{
				"scripts": {
					"start": "some-start-command",
					"poststart": "some-poststart-command"
				}
			}`), 0644)
			Expect(err).NotTo(HaveOccurred())
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
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Launch: packit.LaunchMetadata{
					Processes: []packit.Process{
						{
							Type:    "web",
							Command: fmt.Sprintf("cd %s && some-start-command && some-poststart-command", filepath.Join(workingDir, "some-project-dir")),
						},
					}},
			}))
		})
	})

	context("when the package.json does not include a poststart command", func() {
		it.Before(func() {
			err := ioutil.WriteFile(filepath.Join(workingDir, "some-project-dir", "package.json"), []byte(`{
				"scripts": {
					"prestart": "some-prestart-command",
					"start": "some-start-command"
				}
			}`), 0644)
			Expect(err).NotTo(HaveOccurred())
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
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Launch: packit.LaunchMetadata{
					Processes: []packit.Process{
						{
							Type:    "web",
							Command: fmt.Sprintf("cd %s && some-prestart-command && some-start-command", filepath.Join(workingDir, "some-project-dir")),
						},
					},
				},
			}))
		})
	})

	context("when the package.json does not include a start command", func() {
		it.Before(func() {
			err := ioutil.WriteFile(filepath.Join(workingDir, "some-project-dir", "package.json"), []byte(`{
				"scripts": {
					"prestart": "some-prestart-command",
					"poststart": "some-poststart-command"
				}
			}`), 0644)
			Expect(err).NotTo(HaveOccurred())
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
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Launch: packit.LaunchMetadata{
					Processes: []packit.Process{
						{
							Type:    "web",
							Command: fmt.Sprintf("cd %s && some-prestart-command && node server.js && some-poststart-command", filepath.Join(workingDir, "some-project-dir")),
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
				Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("Unable to open package.json: open %s", filepath.Join(workingDir, "some-project-dir", "package.json")))))
			})
		})

		context("when package.json is malformed", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-project-dir", "package.json"), []byte("%%%"), os.ModePerm)).To(Succeed())
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
				Expect(err).To(MatchError(ContainSubstring("Unable to decode package.json: invalid character")))
			})
		})

		context("when the path parser returns an error", func() {
			it.Before(func() {
				pathParser.GetCall.Returns.Err = errors.New("path-parser-error")
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
				Expect(err).To(MatchError("path-parser-error"))
			})
		})
	})
}
