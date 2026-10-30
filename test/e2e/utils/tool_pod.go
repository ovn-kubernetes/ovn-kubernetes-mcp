// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateToolPod starts the real tools image without its networking daemons.
// The pod is deleted when the calling spec or container finishes.
func CreateToolPod(ctx context.Context, kubeClient client.Client, namespace, image string, env []corev1.EnvVar) string {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{GenerateName: "command-errors-", Namespace: namespace},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{{
				Name: "tools", Image: image, Command: []string{"sleep", "infinity"}, Env: env,
				SecurityContext: RestrictedSecurityContext(),
			}},
		},
	}
	gomega.Expect(kubeClient.Create(ctx, pod)).To(gomega.Succeed())
	ginkgo.DeferCleanup(func(ctx ginkgo.SpecContext) {
		gomega.Expect(client.IgnoreNotFound(kubeClient.Delete(ctx, pod))).To(gomega.Succeed())
	})
	gomega.Eventually(func() bool {
		if err := kubeClient.Get(ctx, client.ObjectKeyFromObject(pod), pod); err != nil {
			return false
		}
		return pod.Status.Phase == corev1.PodRunning
	}, time.Minute, time.Second).Should(gomega.BeTrue())
	return pod.Name
}
