package yarnstart_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit"
	yarnstart "github.com/paketo-buildpacks/yarn-start"
	"github.com/paketo-buildpacks/yarn-start/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir        string
		projectPathParser *fakes.PathParser
		detect            packit.DetectFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Mkdir(filepath.Join(workingDir, "custom"), os.ModePerm)).To(Succeed())

		projectPathParser = &fakes.PathParser{}
		projectPathParser.GetCall.Returns.ProjectPath = "custom"

		detect = yarnstart.Detect(projectPathParser)
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("when there is a yarn.lock", func() {
		it.Before(func() {
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "custom", "yarn.lock"), nil, 0644)).To(Succeed())
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
			Expect(projectPathParser.GetCall.Receives.Path).To(Equal(filepath.Join(workingDir)))
		})
	})

	context("when there is no yarn.lock", func() {
		it("fails detection", func() {
			_, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).To(MatchError(packit.Fail))
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
				Expect(err).To(MatchError(ContainSubstring("failed to stat yarn.lock:")))
			})
		})

		context("when the project path cannot be found", func() {
			it.Before(func() {
				projectPathParser.GetCall.Returns.Err = errors.New("couldn't find directory")
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError("couldn't find directory"))
			})
		})
	})
}
