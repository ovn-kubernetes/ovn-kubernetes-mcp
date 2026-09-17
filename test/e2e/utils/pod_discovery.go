// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var OVNKubeNamespaces = []string{
	"openshift-ovn-kubernetes",
	"ovn-kubernetes",
}

func FindReadyNode(kubeClient client.Client) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	nodeList := &corev1.NodeList{}
	if err := kubeClient.List(ctx, nodeList); err != nil {
		return "", fmt.Errorf("failed to list nodes: %w", err)
	}

	for _, node := range nodeList.Items {
		isReady := false
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				isReady = true
				break
			}
		}
		if !isReady {
			continue
		}
		if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok {
			continue
		}
		if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
			continue
		}
		return node.Name, nil
	}

	for _, node := range nodeList.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				return node.Name, nil
			}
		}
	}

	return "", fmt.Errorf("no ready node found")
}

func FindOVNKubeNodePod(kubeClient client.Client) (namespace, name string, err error) {
	ctx := context.Background()

	var foundNS, foundName string
	var lastErr error

	pollErr := wait.PollUntilContextTimeout(ctx, 2*time.Second, 2*time.Minute, true, func(ctx context.Context) (bool, error) {
		for _, ns := range OVNKubeNamespaces {
			podList := &corev1.PodList{}
			if err := kubeClient.List(ctx, podList,
				client.InNamespace(ns),
				client.MatchingLabels{"app": "ovnkube-node"},
			); err != nil {
				lastErr = err
				continue
			}
			for _, pod := range podList.Items {
				if pod.Status.Phase != corev1.PodRunning {
					continue
				}
				for _, c := range pod.Status.Conditions {
					if c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
						foundNS = ns
						foundName = pod.Name
						return true, nil
					}
				}
			}
		}
		return false, nil
	})
	if pollErr != nil {
		if lastErr != nil {
			return "", "", fmt.Errorf("pod discovery timed out, last error: %w", lastErr)
		}
		return "", "", fmt.Errorf("no running ovnkube-node pod found")
	}
	return foundNS, foundName, nil
}
