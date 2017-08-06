package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"
)

var _ = Describe("Configuration", func() {

	Describe("Discover configuration", func() {

		It("must create the configuration from environment variables", func() {
			api := "https://api.io"
			user := "user"
			password := "password"

			os.Setenv("CF_TARGET", api)
			os.Setenv("NAME", user)
			os.Setenv("PASSWORD", password)

			disc, err := discoverConfig()

			Ω(err).Should(BeNil())
			Ω(disc.ApiAddress).Should(Equal(api))
			Ω(disc.Username).Should(Equal(user))
			Ω(disc.Password).Should(Equal(password))
		})

		It("must fail creating the configuration with unsufficient data", func() {
			api := "https://api.io"
			password := "password"

			os.Setenv("CF_TARGET", api)
			os.Setenv("PASSWORD", password)
			os.Unsetenv("NAME")

			disc, err := discoverConfig()

			Ω(err).ShouldNot(BeNil())
			Ω(disc).Should(BeNil())
		})
	})

})
