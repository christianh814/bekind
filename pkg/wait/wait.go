package wait

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

// IsDeploymentRunning checks to see if the named deployment is running.
func IsDeploymentRunning(c kubernetes.Interface, ns string, depl string) wait.ConditionFunc {
	return func() (bool, error) {
		// Get the named deployment
		dep, err := c.AppsV1().Deployments(ns).Get(context.TODO(), depl, metav1.GetOptions{})

		// If the deployment is not found, that's okay. It means it's not up and running yet
		if errors.IsNotFound(err) {
			return false, nil
		}

		// if another error was found, return that
		if err != nil {
			return false, err
		}

		// If the deployment hasn't finished, then let's run again
		if dep.Status.ReadyReplicas == 0 {
			return false, nil
		}

		return true, nil
	}
}

// WaitForDeployment polls up to timeout seconds for a pod to enter the running state.
func WaitForDeployment(c kubernetes.Interface, namespace string, deployment string, timeout time.Duration) error {
	return wait.PollImmediate(5*time.Second, timeout, IsDeploymentRunning(c, namespace, deployment))
}
