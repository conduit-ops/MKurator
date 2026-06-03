//go:build e2e
// +build e2e

package e2e

import (
	"time"

	. "github.com/onsi/gomega"
)

func eventuallyExpectQMCReady(ns string) {
	Eventually(func(g Gomega) {
		out, err := runKubectl("get", "queuemanagerconnection", mqConnectionName, "-n", ns,
			"-o", "jsonpath={.status.conditions[?(@.type==\"Ready\")].status}")
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(out).To(Equal("True"), "QueueManagerConnection %s should be Ready", mqConnectionName)
	}).WithTimeout(qmcRotationEventuallyTimeout).WithPolling(5 * time.Second).Should(Succeed())
}
