package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"
)

var _ = Describe("Configuration", func() {

	Describe("Discover configuration", func() {

		It("must parse the environment variables correctly", func() {
			os.Setenv("UC_DOCKER_IMAGE_PULL", "true")
			disc, err := discoverConfig()
			Ω(err).Should(BeNil())
			Ω(disc.AllowImagePull).Should(BeTrue())

			os.Setenv("UC_DOCKER_IMAGE_PULL", "false")
			disc, err = discoverConfig()
			Ω(err).Should(BeNil())
			Ω(disc.AllowImagePull).Should(BeFalse())

			os.Setenv("UC_DOCKER_IMAGE_PULL", "XyZ")
			disc, err = discoverConfig()
			Ω(err).ShouldNot(BeNil())
			Ω(disc).Should(BeNil())
		})
	})

})
