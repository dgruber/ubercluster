package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/dgruber/ubercluster/pkg/types"
)

var _ = Describe("Run task", func() {

	Describe("Create task request", func() {

		It("must transform correctly ", func() {
			tr, err := createTaskRequest(types.JobTemplate{
				RemoteCommand: "command",
				JobName:       "name",
				MinPhysMemory: 1024,
			}, "123")
			Expect(err).To(BeNil())
			Expect(tr.Command).To(BeEquivalentTo("command"))
			Expect(tr.Name).To(BeEquivalentTo("name"))
			Expect(tr.MemoryInMegabyte).To(BeEquivalentTo(1))
			Expect(tr.DiskInMegabyte).To(BeEquivalentTo(0))
			Expect(tr.DropletGUID).To(BeEquivalentTo("123"))
		})

		It("must transform args to command", func() {
			tr, err := createTaskRequest(types.JobTemplate{
				RemoteCommand: "command",
				Args:          []string{"1", "2"},
				JobName:       "name",
				MinPhysMemory: 1024,
			}, "123")
			Expect(err).To(BeNil())
			Expect(tr.Command).To(BeEquivalentTo("command 1 2"))
			Expect(tr.Name).To(BeEquivalentTo("name"))
			Expect(tr.MemoryInMegabyte).To(BeEquivalentTo(1))
			Expect(tr.DiskInMegabyte).To(BeEquivalentTo(0))
			Expect(tr.DropletGUID).To(BeEquivalentTo("123"))
		})
	})

})
