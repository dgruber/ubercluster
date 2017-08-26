package main_test

import (
	. "github.com/dgruber/ubercluster/cmd/dockerproxy"
	"github.com/dgruber/ubercluster/cmd/dockerproxy/fake"
	"github.com/dgruber/ubercluster/pkg/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Proxy", func() {

	var (
		config *DockerConfig
		f      *fake.FakeDocker
	)

	BeforeEach(func() {
		config = &DockerConfig{
			AllowImagePull: true,
		}
		f = fake.NewFakeDocker()
	})

	Context("Proxy interface functions", func() {

		It("must display DRMS information", func() {
			p := NewProxy(f, config)
			Ω(p.DRMSVersion()).Should(Equal("1.0.0"))
			Ω(p.DRMSName()).Should(Equal("Docker"))
			Ω(p.DRMSLoad()).Should(BeNumerically("==", 0.5))
		})

		It("must show machines, queues, sessions", func() {
			p := NewProxy(f, config)

			m, err := p.GetAllMachines(nil)
			Ω(err).Should(BeNil())
			Ω(m).ShouldNot(BeNil())

			q, err := p.GetAllQueues(nil)
			Ω(err).Should(BeNil())
			Ω(q).ShouldNot(BeNil())

			s, err := p.GetAllSessions(nil)
			Ω(err).Should(BeNil())
			Ω(s).ShouldNot(BeNil())
		})

		It("must show all images / job categories", func() {
			p := NewProxy(f, config)
			cats, err := p.GetAllCategories()
			Ω(err).Should(BeNil())
			Ω(cats).ShouldNot(BeNil())
		})

		It("should run a job", func() {
			p := NewProxy(f, config)
			cats, err := p.RunJob(types.JobTemplate{JobCategory: "notExisting"})
			Ω(err).Should(BeNil())
			Ω(cats).ShouldNot(BeNil())

			cats, err = p.RunJob(types.JobTemplate{JobCategory: "golang/latest"})
			Ω(err).Should(BeNil())
			Ω(cats).ShouldNot(BeNil())
		})

		It("should perform job operations", func() {
			p := NewProxy(f, config)
			id, err := p.RunJob(types.JobTemplate{JobCategory: "notExisting"})
			Ω(err).Should(BeNil())
			Ω(id).ShouldNot(BeNil())

			out, err := p.JobOperation("", "suspend", id)
			Ω(err).Should(BeNil())
			Ω(out).Should(Equal("Suspended job"))

			out, err = p.JobOperation("", "resume", id)
			Ω(err).Should(BeNil())
			Ω(out).Should(Equal("Resumed job"))

			out, err = p.JobOperation("", "terminate", id)
			Ω(err).Should(BeNil())
			Ω(out).Should(Equal("Terminated job"))

			_, err = p.JobOperation("", "XYZ", id)
			Ω(err).ShouldNot(BeNil())
		})

	})

})
