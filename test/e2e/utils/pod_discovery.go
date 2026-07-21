package utils

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var OVNKubeNamespaces = []string{
	"openshift-ovn-kubernetes",
	"ovn-kubernetes",
}

func FindOVNKubeNodePod(kubeClient client.Client) (namespace, name string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var lastErr error

	for {
		for _, ns := range OVNKubeNamespaces {
			podList := &corev1.PodList{}
			if err := kubeClient.List(ctx, podList,
				client.InNamespace(ns),
				client.MatchingLabels{"app": "ovnkube-node"},
			); err != nil {
				if ctx.Err() == nil {
					lastErr = err
				}
				continue
			}
			for _, pod := range podList.Items {
				if pod.Status.Phase != corev1.PodRunning {
					continue
				}
				for _, c := range pod.Status.Conditions {
					if c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
						return ns, pod.Name, nil
					}
				}
			}
		}

		select {
		case <-ctx.Done():
			if lastErr != nil {
				return "", "", fmt.Errorf("pod discovery timed out, last error: %w", lastErr)
			}
			return "", "", fmt.Errorf("no running ovnkube-node pod found")
		case <-time.After(2 * time.Second):
		}
	}
}
