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
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var (
	buildpack            string
	nodeBuildpack        string
	tiniBuildpack        string
	yarnBuildpack        string
	yarnInstallBuildpack string

	buildpackInfo struct {
		Buildpack struct {
			ID   string
			Name string
		}
	}

	config struct {
		NodeEngine  string `json:"node-engine"`
		Yarn        string `json:"yarn"`
		YarnInstall string `json:"yarn-install"`
		Tini        string `json:"tini"`
	}
)

func TestIntegration(t *testing.T) {
	var (
		Expect = NewWithT(t).Expect
		err    error
	)

	root, err := filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	file, err := os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())

	_, err = toml.DecodeReader(file, &buildpackInfo)
	Expect(err).NotTo(HaveOccurred())
	Expect(file.Close()).To(Succeed())

	file, err = os.Open("../integration.json")
	Expect(err).NotTo(HaveOccurred())

	Expect(json.NewDecoder(file).Decode(&config)).To(Succeed())
	Expect(file.Close()).To(Succeed())

	buildpackStore := occam.NewBuildpackStore()

	buildpack, err = buildpackStore.Get.
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	nodeBuildpack, err = buildpackStore.Get.
		Execute(config.NodeEngine)
	Expect(err).NotTo(HaveOccurred())

	yarnBuildpack, err = buildpackStore.Get.
		Execute(config.Yarn)
	Expect(err).NotTo(HaveOccurred())

	yarnInstallBuildpack, err = buildpackStore.Get.
		Execute(config.YarnInstall)
	Expect(err).NotTo(HaveOccurred())

	tiniBuildpack, err = buildpackStore.Get.
		Execute(config.Tini)
	Expect(err).NotTo(HaveOccurred())

	SetDefaultEventuallyTimeout(5 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Default", testDefault)
	suite.Run(t)
}
