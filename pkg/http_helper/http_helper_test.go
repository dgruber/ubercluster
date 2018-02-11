package http_helper_test

import (
	. "github.com/dgruber/ubercluster/pkg/http_helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("HttpHelper", func() {

	otpRequest := "1234567"

	Context("basic functionality", func() {

		It("should add the one-time-password to GET if present", func() {
			var otp string
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				otp = r.FormValue("otp")
			}))
			defer ts.Close()

			_, err := UberGet(&http.Client{}, otpRequest, ts.URL)
			Ω(err).Should(BeNil())
			Ω(otp).Should(Equal(otpRequest))

			_, err = UberGet(&http.Client{}, otpRequest, ts.URL+"?key=value")
			Ω(err).Should(BeNil())
			Ω(otp).Should(Equal(otpRequest))

			_, err = UberGet(&http.Client{}, "", ts.URL)
			Ω(err).Should(BeNil())
			Ω(otp).Should(Equal(""))
		})

		It("should add the one-time-password to POST", func() {
			var otp string
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				otp = r.FormValue("otp")
			}))
			defer ts.Close()

			_, err := UberPost(&http.Client{}, otpRequest, ts.URL, "", bytes.NewReader(nil))
			Ω(err).Should(BeNil())
			Ω(otp).Should(Equal(otpRequest))

			_, err = UberPost(&http.Client{}, otpRequest, ts.URL+"?key=value", "", bytes.NewReader(nil))
			Ω(err).Should(BeNil())
			Ω(otp).Should(Equal(otpRequest))

			_, err = UberPost(&http.Client{}, "", ts.URL, "", bytes.NewReader(nil))
			Ω(err).Should(BeNil())
			Ω(otp).Should(Equal(""))
		})

	})

})
