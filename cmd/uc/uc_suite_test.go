package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestUc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Uc Suite")
}
