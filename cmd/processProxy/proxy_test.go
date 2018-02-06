package main_test

import (
	. "github.com/dgruber/ubercluster/cmd/processProxy"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/dgruber/ubercluster/pkg/types"
)

var _ = Describe("Proxy", func() {

	jtemplate := types.JobTemplate{RemoteCommand: "sleep", Args: []string{"0"}}

	Context("basic operations", func() {
		proxy := NewProxy()

		It("should be possible to create a NewProxy()", func() {
			Ω(proxy.SessionManager).ShouldNot(BeNil())
			Ω(proxy.JobSession).ShouldNot(BeNil())
		})

		It("should be possible to Run() a job", func() {
			jobid, err := proxy.RunJob(jtemplate)
			Ω(err).Should(BeNil())
			Ω(jobid).Should(Equal("1"))
		})

		It("should be possible to do a JobOperation()", func() {
			jobid, err := proxy.RunJob(jtemplate)
			Ω(err).Should(BeNil())
			_, errOp := proxy.JobOperation(SESSION_NAME, "suspend", jobid)
			Ω(errOp).Should(BeNil())
			_, errOp = proxy.JobOperation(SESSION_NAME, "resume", jobid)
			Ω(errOp).Should(BeNil())
		})

		It("should be possible to GetJobInfosByFilter()", func() {
			jis := proxy.GetJobInfosByFilter(false, types.JobInfo{})
			Ω(jis).ShouldNot(BeNil())
		})

		It("should be possible to GetJobInfo()", func() {
			jobid, err := proxy.RunJob(jtemplate)
			Ω(err).Should(BeNil())
			ji := proxy.GetJobInfo(jobid)
			Ω(ji).ShouldNot(BeNil())
		})

		It("should be possible to GetAllMaschines()", func() {
			hostnames, err := proxy.GetAllMachines(nil)
			Ω(err).Should(BeNil())
			Ω(hostnames).ShouldNot(BeNil())
			Ω(len(hostnames)).Should(BeNumerically("==", 1))
		})

		It("should be possible to filter GetAllMaschines()", func() {
			hostnames, err := proxy.GetAllMachines([]string{"x"})
			Ω(err).Should(BeNil())
			Ω(hostnames).ShouldNot(BeNil())
			Ω(len(hostnames)).Should(BeNumerically("==", 0))
		})

		It("should be possible to GetAllQueues()", func() {
			queues, err := proxy.GetAllQueues(nil)
			Ω(err).Should(BeNil())
			Ω(queues).ShouldNot(BeNil())
			Ω(len(queues)).Should(BeNumerically("==", 1))
			Ω(queues[0].Name).Should(Equal("os"))
		})

		It("should be possible to filter GetAllQueues()", func() {
			queues, err := proxy.GetAllQueues([]string{"x"})
			Ω(err).Should(BeNil())
			Ω(queues).ShouldNot(BeNil())
			Ω(len(queues)).Should(BeNumerically("==", 0))
		})

		It("should be possible to GetAllSessions()", func() {
			jobsessions, err := proxy.GetAllSessions(nil)
			Ω(err).Should(BeNil())
			Ω(jobsessions).ShouldNot(BeNil())
			Ω(len(jobsessions)).Should(BeNumerically("==", 1))
			Ω(jobsessions[0]).Should(Equal(SESSION_NAME))
		})

		It("should be possible to filter GetAllSessions()", func() {
			jobsessions, err := proxy.GetAllSessions([]string{"x"})
			Ω(err).Should(BeNil())
			Ω(jobsessions).ShouldNot(BeNil())
			Ω(len(jobsessions)).Should(BeNumerically("==", 0))
		})

		It("should be possible to get GetAllCategories()", func() {
			cats, err := proxy.GetAllCategories()
			Ω(err).Should(BeNil())
			Ω(cats).ShouldNot(BeNil())
			Ω(len(cats)).Should(BeNumerically("==", 0))
		})

		It("should be possible to get DRMSVersion()", func() {
			version := proxy.DRMSVersion()
			Ω(version).ShouldNot(Equal(""))
		})

		It("should be possible to get DRMSName()", func() {
			name := proxy.DRMSName()
			Ω(name).ShouldNot(Equal(""))
		})

		It("should be possible to get DRMSLoad()", func() {
			load := proxy.DRMSLoad()
			Ω(load).ShouldNot(BeNumerically("==", 0.0))
		})

	})

})
