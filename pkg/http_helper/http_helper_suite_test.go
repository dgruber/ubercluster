package http_helper_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHttpHelper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HttpHelper Suite")
}
