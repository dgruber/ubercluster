package main_test

import (
	. "github.com/dgruber/ubercluster/cmd/processProxy"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/ubercluster/pkg/types"
)

var _ = Describe("ConvertJobInfo", func() {

	Context("Check fields of converted JobInfo", func() {

		var input drmaa2interface.JobInfo
		var expected types.JobInfo

		BeforeEach(func() {
			input = drmaa2interface.JobInfo{
				ID:                "77",
				ExitStatus:        13,
				TerminatingSignal: "terminatingSignal",
				Annotation:        "annotation",
				State:             drmaa2interface.Suspended,
				SubState:          "subState",
				AllocatedMachines: []string{"machine1", "machine2"},
				SubmissionMachine: "localhost",
				JobOwner:          "owner",
				Slots:             1,
				QueueName:         "queue",
				WallclockTime:     time.Hour,
				CPUTime:           1000,
				SubmissionTime:    time.Unix(64000000, 0),
				DispatchTime:      time.Unix(65000000, 0),
				FinishTime:        time.Unix(66000000, 0),
			}

			expected = types.JobInfo{
				Id:                "77",
				ExitStatus:        13,
				TerminatingSignal: "terminatingSignal",
				Annotation:        "annotation",
				State:             types.Suspended,
				SubState:          "subState",
				AllocatedMachines: []string{"machine1", "machine2"},
				SubmissionMachine: "localhost",
				JobOwner:          "owner",
				Slots:             1,
				QueueName:         "queue",
				WallclockTime:     time.Hour,
				CPUTime:           1000,
				SubmissionTime:    time.Unix(64000000, 0),
				DispatchTime:      time.Unix(65000000, 0),
				FinishTime:        time.Unix(66000000, 0),
			}
		})

		It("must contain all of them", func() {
			output := ConvertJobInfo(input)
			Ω(output.Id).Should(Equal(expected.Id))
			Ω(output.ExitStatus).Should(Equal(expected.ExitStatus))
			Ω(output.TerminatingSignal).Should(Equal(expected.TerminatingSignal))
			Ω(output.Annotation).Should(Equal(expected.Annotation))
			Ω(output.State).Should(Equal(expected.State))
			Ω(output.SubState).Should(Equal(expected.SubState))
			Ω(output.AllocatedMachines).Should(Equal(expected.AllocatedMachines))
			Ω(output.SubmissionMachine).Should(Equal(expected.SubmissionMachine))
			Ω(output.JobOwner).Should(Equal(expected.JobOwner))
			Ω(output.Slots).Should(Equal(expected.Slots))
			Ω(output.QueueName).Should(Equal(expected.QueueName))
			Ω(output.WallclockTime).Should(Equal(expected.WallclockTime))
			Ω(output.CPUTime).Should(Equal(expected.CPUTime))
			Ω(output.SubmissionTime).Should(Equal(expected.SubmissionTime))
			Ω(output.DispatchTime).Should(Equal(expected.DispatchTime))
			Ω(output.FinishTime).Should(Equal(expected.FinishTime))
		})

	})

})
