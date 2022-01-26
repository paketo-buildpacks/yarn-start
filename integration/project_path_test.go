package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testProjectPath(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when building an app with a custom project path set", func() {
		var (
			image     occam.Image
			container occam.Container

			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("builds a working OCI image and runs given start cmd", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "custom_project_path_app"))
			Expect(err).NotTo(HaveOccurred())

			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithBuildpacks(
					settings.Buildpacks.NodeEngine.Online,
					settings.Buildpacks.Yarn.Online,
					settings.Buildpacks.YarnInstall.Online,
					settings.Buildpacks.YarnStart.Online,
				).
				WithPullPolicy("never").
				WithEnv(map[string]string{"BP_NODE_PROJECT_PATH": "hello_world_server"}).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs.String())

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
				"  Assigning launch processes:",
				`    web (default): bash -c cd /workspace/hello_world_server && echo "prehello" && echo "starthello" && node server.js && echo "posthello"`,
				"",
			))

			container, err = docker.Container.Run.
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				WithPublishAll().
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(container).Should(BeAvailable())
			Eventually(container).Should(Serve(ContainSubstring("Hello, World!")))

			cLogs := func() string {
				containerLogs, err := docker.Container.Logs.Execute(container.ID)
				Expect(err).NotTo(HaveOccurred())
				return containerLogs.String()
			}

			Eventually(cLogs).Should(ContainSubstring("prehello"))
			Eventually(cLogs).Should(ContainSubstring("starthello"))
		})

		context("when BP_LIVE_RELOAD_ENABLED=true during the build", func() {
			it("makes the default process reloadable and watches the correct subdirectory", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "custom_project_path_app"))
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithBuildpacks(
						settings.Buildpacks.Watchexec.Online,
						settings.Buildpacks.NodeEngine.Online,
						settings.Buildpacks.Yarn.Online,
						settings.Buildpacks.YarnInstall.Online,
						settings.Buildpacks.YarnStart.Online,
					).
					WithPullPolicy("never").
					WithEnv(map[string]string{
						"BP_NODE_PROJECT_PATH":   "hello_world_server",
						"BP_LIVE_RELOAD_ENABLED": "true",
					}).
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred(), logs.String())

				Expect(logs).To(ContainLines(
					MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
					"  Assigning launch processes:",
					`    web (default): watchexec --restart --shell none --watch /workspace/hello_world_server --ignore /workspace/hello_world_server/package.json --ignore /workspace/hello_world_server/yarn.lock --ignore /workspace/hello_world_server/node_modules -- bash -c cd /workspace/hello_world_server && echo "prehello" && echo "starthello" && node server.js && echo "posthello"`,
					`    no-reload:     bash -c cd /workspace/hello_world_server && echo "prehello" && echo "starthello" && node server.js && echo "posthello"`,
					"",
				))

				container, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(BeAvailable())
				Eventually(container).Should(Serve(ContainSubstring("Hello, World!")))

				cLogs := func() string {
					containerLogs, err := docker.Container.Logs.Execute(container.ID)
					Expect(err).NotTo(HaveOccurred())
					return containerLogs.String()
				}

				Eventually(cLogs).Should(ContainSubstring("prehello"))
				Eventually(cLogs).Should(ContainSubstring("starthello"))
			})
		})
	})
}
