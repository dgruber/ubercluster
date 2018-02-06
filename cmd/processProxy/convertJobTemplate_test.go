package main_test

import (
	. "github.com/dgruber/ubercluster/cmd/processProxy"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/dgruber/ubercluster/pkg/types"
)

var _ = Describe("ConvertJobTemplate", func() {

	Context("Basic operation", func() {
		var input types.JobTemplate

		BeforeEach(func() {
			input = types.JobTemplate{
				RemoteCommand: "command",
			}
		})

		It("should convert the job template", func() {
			output := ConvertJobTemplate(input)
			Ω(output.RemoteCommand).Should(Equal(input.RemoteCommand))
		})
	})

})
