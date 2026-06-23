package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
)

func (c *OVNKMCPServerClientSet) DebugNode(ctx context.Context, namespace, name, image string, command []string, hostPath, mountPath string, timeout time.Duration) (string, string, error) {
	// Validate name
	if name == "" {
		return "", "", fmt.Errorf("node name is required")
	}

	// Validate paths before creating the pod
	if err := utils.ValidatePath(hostPath, "hostPath", true); err != nil {
		return "", "", err
	}

	if err := utils.ValidatePath(mountPath, "mountPath", true); err != nil {
		return "", "", err
	}

	// If namespace is not provided, use the default namespace
	if namespace == "" {
		namespace = metav1.NamespaceDefault
	}

	// If timeout is specified, create a new context with timeout
	if timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	debugPodName, cleanupPod, err := c.createPod(ctx, name, namespace, image, hostPath, mountPath)
	if err != nil {
		return "", "", err
	}

	if cleanupPod != nil {
		defer cleanupPod()
	}

	// Execute the command in the pod.
	stdout, stderr, err := c.ExecPod(ctx, debugPodName, namespace, "debug-container", command)
	if err != nil {
		return "", "", fmt.Errorf("failed to execute command in debug pod: %w", err)
	}

	return stdout, stderr, nil
}

func (c *OVNKMCPServerClientSet) createPod(ctx context.Context, node, namespace, image, hostPath, mountPath string) (string, func(), error) {
	hostPathType := corev1.HostPathDirectory
	sleepCommand := []string{"sleep", "infinity"}

	if hostPath == "" {
		hostPath = "/"
	}

	if mountPath == "" {
		mountPath = "/host"
	}

	var envVars []corev1.EnvVar
	if hostPath == "/" {
		// to collect sos report requires this env var is set when hostPath is /
		envVars = []corev1.EnvVar{
			{
				Name:  "HOST",
				Value: mountPath,
			},
		}
	}

	// Create a host networked privileged debug pod.
	debugPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "debug-node-" + node + "-",
			Namespace:    namespace,
		},
		Spec: corev1.PodSpec{
			NodeName:      node,
			RestartPolicy: corev1.RestartPolicyNever,
			Tolerations: []corev1.Toleration{
				{
					Operator: corev1.TolerationOpExists,
				},
			},
			HostNetwork: true,
			HostPID:     true,
			HostIPC:     true,
			Volumes: []corev1.Volume{
				{
					Name: "host",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: hostPath,
							Type: &hostPathType,
						},
					},
				},
			},
			Containers: []corev1.Container{
				{
					Name:    "debug-container",
					Image:   image,
					Command: sleepCommand,
					SecurityContext: &corev1.SecurityContext{
						Privileged: ptr.To(true),
						RunAsUser:  ptr.To(int64(0)),
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "host",
							MountPath: mountPath,
						},
					},
					Env: envVars,
				}},
		},
	}

	// Create the debug pod.
	createdDebugPod, err := c.clientSet.CoreV1().Pods(namespace).Create(ctx, debugPod, metav1.CreateOptions{})
	if err != nil {
		return "", nil, fmt.Errorf("failed to create debug pod: %w", err)
	}

	cleanupPod := func() {
		// Create a new context with a timeout of 10 seconds to delete the pod.
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Delete the pod.
		err := c.clientSet.CoreV1().Pods(namespace).Delete(cleanupCtx, createdDebugPod.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Printf("failed to cleanup debug pod: %v", err)
		}
	}

	// Wait for the pod to be running.
	err = wait.PollUntilContextTimeout(ctx, time.Millisecond*500, time.Minute*1, true, func(ctx context.Context) (bool, error) {
		pod, err := c.clientSet.CoreV1().Pods(namespace).Get(ctx, createdDebugPod.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		// Return true if the pod is running.
		return pod.Status.Phase == corev1.PodRunning, nil
	})
	if err != nil {
		cleanupPod()
		return "", nil, fmt.Errorf("debug pod did not reach running state within timeout of 1 minute: %w", err)
	}

	return createdDebugPod.Name, cleanupPod, nil
}
