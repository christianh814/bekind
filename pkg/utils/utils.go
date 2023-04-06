package utils

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	goyaml "gopkg.in/yaml.v2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

func NewClient(kubeConfigPath string) (kubernetes.Interface, error) {
	if kubeConfigPath == "" {
		kubeConfigPath = os.Getenv("KUBECONFIG")
	}
	if kubeConfigPath == "" {
		kubeConfigPath = clientcmd.RecommendedHomeFile // use default path(.kube/config)
	}
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(kubeConfig)
}

func DownloadFileString(url string) (string, error) {
	r, err := http.Get(url)
	if err != nil {
		return "", err
	}
	buf := new(strings.Builder)
	_, err = io.Copy(buf, r.Body)
	return buf.String(), err
}

func SplitYAML(resources []byte) ([][]byte, error) {
	dec := goyaml.NewDecoder(bytes.NewReader(resources))

	var res [][]byte
	for {
		var value interface{}
		err := dec.Decode(&value)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		valueBytes, err := goyaml.Marshal(value)
		if err != nil {
			return nil, err
		}
		res = append(res, valueBytes)
	}
	return res, nil
}

func WaitForDeployment(c kubernetes.Interface, namespace string, deployment string, timeout time.Duration) error {
	return wait.PollImmediate(5*time.Second, timeout, IsDeploymentRunning(c, namespace, deployment))
}

func IsDeploymentRunning(c kubernetes.Interface, ns string, depl string) wait.ConditionFunc {
	return func() (bool, error) {
		dep, err := c.AppsV1().Deployments(ns).Get(context.TODO(), depl, v1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if dep.Status.ReadyReplicas == 0 {
			return false, nil
		}
		return true, nil
	}
}

func LabelWorkers(c kubernetes.Interface) error {
	workers, err := c.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{
		LabelSelector: `!node-role.kubernetes.io/control-plane`,
	})
	if err != nil {
		return err
	}

	for _, w := range workers.Items {
		labelKey := "node-role.kubernetes.io/worker"
		labelValue := ""

		labels := w.Labels
		labels[labelKey] = labelValue
		w.SetLabels(labels)

		_, err = c.CoreV1().Nodes().Update(context.TODO(), &w, v1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}
