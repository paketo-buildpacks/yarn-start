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

func testCustomStartCmd(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker

		pullPolicy       = "never"
		extenderBuildStr = ""
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()

		if settings.Extensions.UbiNodejsExtension.Online != "" {
			pullPolicy = "always"
			extenderBuildStr = "[extender (build)] "
		}
	})

	context("when building a container image with pack", func() {
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
			source, err = occam.Source(filepath.Join("testdata", "custom_start_cmd_app"))
			Expect(err).NotTo(HaveOccurred())

			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithExtensions(
					settings.Extensions.UbiNodejsExtension.Online,
				).
				WithBuildpacks(
					settings.Buildpacks.NodeEngine.Online,
					settings.Buildpacks.Yarn.Online,
					settings.Buildpacks.YarnInstall.Online,
					settings.Buildpacks.YarnStart.Online,
				).
				WithPullPolicy(pullPolicy).
				Execute(name, source)

			Expect(err).NotTo(HaveOccurred(), logs.String())

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s%s \d+\.\d+\.\d+`, extenderBuildStr, settings.Buildpack.Name)),
			))
			Expect(logs).To(ContainLines(
				extenderBuildStr+"  Assigning launch processes:",
				extenderBuildStr+`    web (default): bash -c echo "prestart" && echo "start" && node server.js && echo "poststart"`,
				extenderBuildStr+"",
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

			Eventually(cLogs).Should(ContainSubstring("prestart"))
			Eventually(cLogs).Should(ContainSubstring("start"))
		})

		context("when BP_LIVE_RELOAD_ENABLED=true during the build", func() {
			it("builds an OCI image that has a reloadable default process and a non-reload process", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "custom_start_cmd_app"))
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithExtensions(
						settings.Extensions.UbiNodejsExtension.Online,
					).
					WithBuildpacks(
						settings.Buildpacks.Watchexec.Online,
						settings.Buildpacks.NodeEngine.Online,
						settings.Buildpacks.Yarn.Online,
						settings.Buildpacks.YarnInstall.Online,
						settings.Buildpacks.YarnStart.Online,
					).
					WithEnv(map[string]string{
						"BP_LIVE_RELOAD_ENABLED": "true",
					}).
					WithPullPolicy(pullPolicy).
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred(), logs.String())

				Expect(logs).To(ContainLines(
					MatchRegexp(fmt.Sprintf(`%s%s \d+\.\d+\.\d+`, extenderBuildStr, settings.Buildpack.Name))))
				Expect(logs).To(ContainLines(
					extenderBuildStr+"  Assigning launch processes:",
					extenderBuildStr+`    web (default): watchexec --restart --shell none --watch /workspace --ignore /workspace/package.json --ignore /workspace/yarn.lock --ignore /workspace/node_modules -- bash -c echo "prestart" && echo "start" && node server.js && echo "poststart"`,
					extenderBuildStr+`    no-reload:     bash -c echo "prestart" && echo "start" && node server.js && echo "poststart"`,
					extenderBuildStr+"",
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

				Eventually(cLogs).Should(ContainSubstring("prestart"))
				Eventually(cLogs).Should(ContainSubstring("start"))
			})
		})
	})
}
