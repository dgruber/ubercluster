package proxy_test

import (
	. "github.com/dgruber/ubercluster/pkg/proxy"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProxySecurity", func() {

	Context("basic functions", func() {

		It("should read the trusted client certs from directory", func() {
			pool, err := ReadTrustedClientCertPool("./testClientCerts")
			Ω(err).Should(BeNil())
			Ω(pool).ShouldNot(BeNil())
		})

	})

	Context("error cases", func() {

		It("fail when directory does not exist", func() {
			pool, err := ReadTrustedClientCertPool("./unknownDir")
			Ω(err).ShouldNot(BeNil())
			Ω(pool).Should(BeNil())
		})

	})

})
