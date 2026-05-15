package client

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"
	"k8s.io/client-go/tools/remotecommand"
)

func TestGetPodLogs(t *testing.T) {
	fakeclient := NewFakeClient()
	logs, err := fakeclient.GetPodLogs(context.Background(), "test", "default", "", false)
	if err != nil {
		t.Fatalf("Failed to get pod logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("Failed to get pod logs: %v", logs)
	}
	if logs[0] != "fake logs" {
		t.Fatalf("Failed to get pod logs: %v", logs)
	}
}

type fakeCommand string

const (
	successExecCommand fakeCommand = "success"
	failureExecCommand fakeCommand = "failure"
	errorExecCommand   fakeCommand = "error"
)

type fakeExecutor struct {
	url     *url.URL
	execErr error
}

func (e *fakeExecutor) Execute(url *url.URL, config *rest.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool, terminalSizeQueue remotecommand.TerminalSizeQueue) error {
	return e.ExecuteWithContext(context.Background(), url, config, stdin, stdout, stderr, tty, nil)
}

func (e *fakeExecutor) ExecuteWithContext(ctx context.Context, url *url.URL, config *rest.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool, terminalSizeQueue remotecommand.TerminalSizeQueue) error {
	e.url = url
	command := url.Query().Get("command")
	switch command {
	case string(successExecCommand):
		_, e.execErr = stdout.Write([]byte("Successfully executed command"))
	case string(failureExecCommand):
		_, e.execErr = stderr.Write([]byte("Failed to execute command"))
	case string(errorExecCommand):
		e.execErr = fmt.Errorf("exec error")
	}

	return e.execErr
}

func TestExecPod(t *testing.T) {
	tests := []struct {
		name          string
		execPath      string
		containerName string
		command       string
		execErr       bool
		pod           *corev1.Pod
	}{
		{
			name:     "Successful pod exec without container name",
			execPath: "/api/v1/namespaces/default/pods/test/exec",
			command:  string(successExecCommand),
			execErr:  false,
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "test",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			},
		},
		{
			name:          "Successful pod exec with container name",
			execPath:      "/api/v1/namespaces/default/pods/test/exec",
			command:       string(successExecCommand),
			execErr:       false,
			containerName: "test",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			},
		},
		{
			name:     "Failed pod exec",
			execPath: "/api/v1/namespaces/default/pods/test/exec",
			command:  string(failureExecCommand),
			execErr:  true,
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "test",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodSucceeded,
				},
			},
		},
		{
			name:     "Pod exec error as pod is succeeded",
			execPath: "/api/v1/namespaces/default/pods/test/exec",
			execErr:  true,
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "test",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodSucceeded,
				},
			},
		},
		{
			name:     "Pod exec error as pod is failed",
			execPath: "/api/v1/namespaces/default/pods/test/exec",
			execErr:  true,
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "test",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodFailed,
				},
			},
		},
	}
	for _, test := range tests {
		fakeclient := NewFakeClient(test.pod)

		fakeclient.corev1RestClient = &fake.RESTClient{
			VersionedAPIPath: "/api/v1",
			GroupVersion:     schema.GroupVersion{Group: "", Version: "v1"},
		}

		ex := &fakeExecutor{}
		fakeclient.podExecutor = ex

		stdout, stderr, err := fakeclient.ExecPod(context.Background(), test.pod.Name, test.pod.Namespace, test.containerName, []string{test.command})

		if err != nil {
			// Verify that the error is expected.
			if !test.execErr {
				t.Fatalf("Unexpected pod exec error: %v", err)
			}
			// Verify that the error is expected for a pod that is not running.
			if test.pod.Status.Phase != corev1.PodRunning {
				expectedErr := fmt.Errorf("cannot exec and run command %v in a container in a pod that is not running; current phase is %s", []string{test.command}, test.pod.Status.Phase)
				if err.Error() != expectedErr.Error() {
					t.Fatalf("Expected pod exec error with command %s: %v, got %v", test.command, expectedErr, err)
				}
			}
			// Verify that the error is expected for an error exec command.
			if test.command == string(errorExecCommand) && err.Error() != "exec error" {
				t.Fatalf("Expected exec error with command %s: %v, got %v", test.command, "exec error", err)
			}
		} else {
			// Verify that the url query contains the container name.
			if test.containerName == "" && !strings.Contains(ex.url.RawQuery, test.pod.Spec.Containers[0].Name) {
				t.Fatalf("Url query does not contain container name: %v", ex.url.RawQuery)
			}
			if test.containerName != "" && !strings.Contains(ex.url.RawQuery, test.containerName) {
				t.Fatalf("Url query does not contain container name: %v", ex.url.RawQuery)
			}
			// Verify that the error is not expected.
			if test.execErr {
				t.Fatalf("Expected pod exec error, got nil")
			}
			// Verify that the stdout is expected.
			if test.command == string(successExecCommand) && stdout != "Successfully executed command" {
				t.Fatalf("Expected stdout with command: %s, got %v", test.command, stdout)
			}
			// Verify that the stderr is expected.
			if test.command == string(failureExecCommand) && stderr != "Failed to execute command" {
				t.Fatalf("Expected stderr with command: %s, got %v", test.command, stderr)
			}
		}

	}
}
