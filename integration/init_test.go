package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/occam"
	"github.com/paketo-buildpacks/occam/packagers"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var settings struct {
	Buildpacks struct {
		NodeEngine struct {
			Online string
		}
		Yarn struct {
			Online string
		}
		YarnInstall struct {
			Online string
		}
		YarnStart struct {
			Online string
		}
		Watchexec struct {
			Online string
		}
	}
	Extensions struct {
		UbiNodejsExtension struct {
			Online string
		}
	}
	Buildpack struct {
		ID   string
		Name string
	}
	Config struct {
		NodeEngine         string `json:"node-engine"`
		Yarn               string `json:"yarn"`
		YarnInstall        string `json:"yarn-install"`
		Watchexec          string `json:"watchexec"`
		UbiNodejsExtension string `json:"ubi-nodejs-extension"`
	}
}

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect
	SetDefaultEventuallyTimeout(10 * time.Second)

	root, err := filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	file, err := os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())

	_, err = toml.NewDecoder(file).Decode(&settings.Buildpack)
	Expect(err).NotTo(HaveOccurred())
	Expect(file.Close()).To(Succeed())

	file, err = os.Open("../integration.json")
	Expect(err).NotTo(HaveOccurred())

	Expect(json.NewDecoder(file).Decode(&settings.Config)).To(Succeed())
	Expect(file.Close()).To(Succeed())

	buildpackStore := occam.NewBuildpackStore()

	pack := occam.NewPack()

	builder, err := pack.Builder.Inspect.Execute()
	Expect(err).NotTo(HaveOccurred())

	if builder.BuilderName == "paketobuildpacks/builder-ubi8-buildpackless-base:latest" {
		settings.Extensions.UbiNodejsExtension.Online, err = buildpackStore.Get.
			Execute(settings.Config.UbiNodejsExtension)
		Expect(err).ToNot(HaveOccurred())
	}

	libpakBuildpackStore := occam.NewBuildpackStore().WithPackager(packagers.NewLibpakTools())

	settings.Buildpacks.YarnStart.Online, err = buildpackStore.Get.
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.NodeEngine.Online, err = buildpackStore.Get.
		Execute(settings.Config.NodeEngine)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.Yarn.Online, err = buildpackStore.Get.
		Execute(settings.Config.Yarn)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.YarnInstall.Online, err = buildpackStore.Get.
		Execute(settings.Config.YarnInstall)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.Watchexec.Online, err = libpakBuildpackStore.Get.
		Execute(settings.Config.Watchexec)
	Expect(err).NotTo(HaveOccurred())

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("CustomStartCmd", testCustomStartCmd)
	suite("Default", testDefault)
	suite("GracefulShutdown", testGracefulShutdown)
	suite("ProjectPath", testProjectPath)
	suite("Workspaces", testWorkspaces)
	suite.Run(t)
}
