package main_test

import (
	. "github.com/dgruber/ubercluster/cmd/uc"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Context("When config.json exists", func() {
		It("must read its two values", func() {
			config := ReadConfig()
			Ω(config.Cluster).To(HaveLen(2))
		})
		It("must have the expected values", func() {
			config := ReadConfig()
			Ω(config.Cluster[1].Name).To(Equal("linux"))
			Ω(config.Cluster[1].Address).To(Equal("http://localhost:1212/"))
			Ω(config.Cluster[1].ProtocolVersion).To(Equal("v1"))
			Ω(config.Cluster[0].Name).To(Equal("default"))
			Ω(config.Cluster[0].Address).To(Equal("http://localhost:8888/"))
			Ω(config.Cluster[0].ProtocolVersion).To(Equal("v1"))
		})
		It("must select the right address", func() {
			clusteraddress, cluster, err := GetClusterAddress("linux")
			Ω(clusteraddress).To(Equal("http://localhost:1212/v1"))
			Ω(cluster).To(Equal("linux"))
			Ω(err).To(BeNil())
			_, _, err2 := GetClusterAddress("liNux")
			Ω(err2).NotTo(BeNil())
		})
	})
})
