package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/dgruber/go-cfclient"
	"time"
)

var _ = Describe("Taskinfo", func() {

	Describe("Transform Tasks in JobInfo", func() {
		It("should tranform correctly", func() {
			adate := time.Date(2016, 12, 22, 00, 00, 00, 00, time.Local)
			task := []cfclient.Task{
				{
					State:     "SUCCEEDED",
					GUID:      "1234",
					CreatedAt: adate,
					UpdatedAt: adate,
					Name:      "Name",
				},
			}
			ji := TransformTasksInJobInfo(task)
			Expect(ji).NotTo(BeNil())
			Expect(len(ji)).To(BeIdenticalTo(1))

			// values
			Ω(ji[0].Id).Should(BeIdenticalTo("1234"))
			Ω(ji[0].SubmissionTime).Should(BeEquivalentTo(adate))
			Ω(ji[0].FinishTime).Should(BeEquivalentTo(adate))
		})
	})

})
