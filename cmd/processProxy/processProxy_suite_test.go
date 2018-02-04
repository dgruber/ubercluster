package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestProcessProxy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ProcessProxy Suite")
}
