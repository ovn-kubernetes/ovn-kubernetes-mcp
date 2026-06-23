package client

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

// GetPodLogs gets the logs of a pod by name and namespace.
func (c *OVNKMCPServerClientSet) GetPodLogs(ctx context.Context, namespace string, name string, container string, previous bool) ([]string, error) {
	req := c.clientSet.CoreV1().Pods(namespace).GetLogs(name,
		&corev1.PodLogOptions{
			Container:  container,
			Timestamps: true,
			Previous:   previous,
		},
	)

	res := req.Do(ctx)
	if res.Error() != nil {
		return nil, fmt.Errorf("failed to fetch pod logs: %w", res.Error())
	}

	logs, err := res.Raw()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pod logs: %w", err)
	}

	return strings.Split(string(logs), "\n"), nil
}

// ExecPod executes a command in a pod by name and namespace.
func (c *OVNKMCPServerClientSet) ExecPod(ctx context.Context, name, namespace, container string, command []string) (string, string, error) {
	pod, err := c.clientSet.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch pod: %w", err)
	}

	// Cannot exec into a container in a pod that is not running.
	if pod.Status.Phase != corev1.PodRunning {
		return "", "", fmt.Errorf("cannot exec and run command %v in a container in a pod that is not running; current phase is %s", command, pod.Status.Phase)
	}

	// If no container is specified, use the first container.
	if container == "" {
		container = pod.Spec.Containers[0].Name
	}

	// Create a request to execute the command.
	req := c.corev1RestClient.
		Post().
		Resource("pods").
		Name(name).
		Namespace(namespace).
		SubResource("exec")

	// Set the parameters for the request.
	req.VersionedParams(&corev1.PodExecOptions{
		Container: container,
		Command:   command,
		Stdout:    true,
		Stderr:    true,
	}, scheme.ParameterCodec)

	// Create buffers for the stdout and stderr.
	stdout := bytes.NewBuffer(make([]byte, 0))
	stderr := bytes.NewBuffer(make([]byte, 0))

	err = c.podExecutor.ExecuteWithContext(ctx, req.URL(), c.config, nil, stdout, stderr, false, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to execute command %v in pod: %w", command, err)
	}

	return stdout.String(), stderr.String(), nil
}
