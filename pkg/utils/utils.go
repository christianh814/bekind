package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	goyaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/pkg/cluster"
)

var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

// GetDefault selected the default runtime from the environment override
func GetDefaultRuntime() cluster.ProviderOption {
	switch p := os.Getenv("KIND_EXPERIMENTAL_PROVIDER"); p {
	case "":
		return nil
	case "podman":
		log.Warn("using podman due to KIND_EXPERIMENTAL_PROVIDER")
		return cluster.ProviderWithPodman()
	case "docker":
		log.Warn("using docker due to KIND_EXPERIMENTAL_PROVIDER")
		return cluster.ProviderWithDocker()
	default:
		log.Warnf("ignoring unknown value %q for KIND_EXPERIMENTAL_PROVIDER", p)
		return nil
	}
}

// DoSSA  does service side apply with the given YAML as a []byte
func DoSSA(ctx context.Context, cfg *rest.Config, yaml []byte) error {
	// Read yaml into a slice of byte
	yml := yaml

	// get the RESTMapper for the GVR
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	// create dymanic client
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}

	// read YAML manifest into unstructured.Unstructured
	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode(yml, nil, obj)
	if err != nil {
		return err
	}

	// Get the GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}

	// Get the REST interface for the GVR
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		dr = dyn.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		// for cluster-wide resources
		dr = dyn.Resource(mapping.Resource)
	}

	// Create object into JSON
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	// Create or Update the obj with service side apply
	//     types.ApplyPatchType indicates service side apply
	//     FieldManager specifies the field owner ID.
	_, err = dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, v1.PatchOptions{
		FieldManager: "bekind",
	})

	return err
}

// check to see if the named deployment is running
// func IsDeploymentRunning(c kubernetes.Interface, ns string, depl string) wait.ConditionFunc {
func IsDeploymentRunning(c kubernetes.Interface, ns string, depl string) wait.ConditionWithContextFunc {

	return func(context.Context) (bool, error) {

		// Get the named deployment
		dep, err := c.AppsV1().Deployments(ns).Get(context.TODO(), depl, v1.GetOptions{})

		// If the deployment is not found, that's okay. It means it's not up and running yet
		if apierrors.IsNotFound(err) {
			return false, nil
		}

		// if another error was found, return that
		if err != nil {
			return false, err
		}

		// If the deployment hasn't finsihed, then let's run again
		if dep.Status.ReadyReplicas == 0 {
			return false, nil
		}

		return true, nil

	}
}

// Poll up to timeout seconds for pod to enter running state.
func WaitForDeployment(c kubernetes.Interface, namespace string, deployment string, timeout time.Duration) error {
	// return wait.PollImmediate(5*time.Second, timeout, IsDeploymentRunning(c, namespace, deployment))
	immediate := true
	return wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, timeout, immediate, IsDeploymentRunning(c, namespace, deployment))
}

// NewClient returns a kubernetes.Interface
func NewClient(kubeConfigPath string) (kubernetes.Interface, error) {
	kubeConfig, err := GetRestConfig(kubeConfigPath)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(kubeConfig)
}

// GetRestConfig returns a *rest.Config
func GetRestConfig(kubeConfigPath string) (*rest.Config, error) {
	if kubeConfigPath == "" {
		kubeConfigPath = os.Getenv("KUBECONFIG")
	}
	if kubeConfigPath == "" {
		kubeConfigPath = clientcmd.RecommendedHomeFile // use default path(.kube/config)
	}
	return clientcmd.BuildConfigFromFlags("", kubeConfigPath)
}

// DownloadFileString will load the contents of a url to a string and return it
func DownloadFileString(url string) (string, error) {
	// Get the data
	r, err := http.Get(url)
	if err != nil {
		return "", err
	}
	//Create a new buffer
	buf := new(strings.Builder)

	// Write the body to file
	_, err = io.Copy(buf, r.Body)
	return buf.String(), err
}

// SplitYAML splits a multipart YAML and returns a slice of a slice of byte
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

// LabelWorkers will label the workers nodes as such
func LabelWorkers(c kubernetes.Interface) error {
	// First select the non control-plane nodes
	workers, err := c.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{
		LabelSelector: `!node-role.kubernetes.io/control-plane`,
	})
	if err != nil {
		return err
	}

	// Loop through and label these nodes as workers
	for _, w := range workers.Items {
		// set up the key and value for the worker
		labelKey := "node-role.kubernetes.io/worker"
		labelValue := ""

		// Apply the labels on the Node object
		labels := w.Labels
		labels[labelKey] = labelValue
		w.SetLabels(labels)

		// Tell the API to update the node
		_, err = c.CoreV1().Nodes().Update(context.TODO(), &w, v1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	// If we made it this far, then we're good
	return nil
}

// ConvertHelmValsToMap converts a slice of strings to a map of strings
func ConvertHelmValsToMap(a []struct {
	Name  string
	Value string
}) map[string]string {
	HelmArgs := map[string]string{
		"set": "",
	}

	if len(a) == 0 {
		return HelmArgs
	} else {

		for i, v := range a {
			if len(a) == i+1 {
				HelmArgs["set"] = HelmArgs["set"] + v.Name + "=" + v.Value
			} else {
				HelmArgs["set"] = HelmArgs["set"] + v.Name + "=" + v.Value + ","

			}
		}

	}

	return HelmArgs

}

// PostInstallManifests will install the manifests after cluster has been created and setup. It is currently best effort/garbage in garbage out
func PostInstallManifests(manifests []string, ctx context.Context, cfg *rest.Config) error {
	// Loop through the manifests and apply them
	for _, m := range manifests {
		// Get the bytes from the manifest
		data, err := getPostInstallBytes(m)
		if err != nil {
			return err
		}

		// Split the YAML into a slice of bytes
		yamls, err := SplitYAML(data)
		if err != nil {
			return err
		}

		// Loop through the yamls and apply them
		for _, y := range yamls {
			// Apply the YAML
			err := DoSSA(ctx, cfg, y)
			if err != nil {
				return err
			}
			// Add 3 second jitter
			//time.Sleep(3 * time.Second)
		}

	}
	// If we are here, then we should be okay
	return nil
}

// SaveBeKindConfig saves the bekind config to a Kubernetes secret
func SaveBeKindConfig(cfg *rest.Config, ctx context.Context, ns string, name string) error {
	// Get the Byteslice of the config
	bekindconfigByteSlice, err := goyaml.Marshal(viper.AllSettings())
	if err != nil {
		return err
	}

	// Create Kubernetes cilent
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}

	// Set up the secret
	secret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Data: map[string][]byte{
			"config.yaml": bekindconfigByteSlice,
		},
		Type: corev1.SecretTypeOpaque, // Default secret type
	}

	// Create the secret
	_, err = client.CoreV1().Secrets(ns).Create(ctx, secret, v1.CreateOptions{})
	if err != nil {
		return err
	}

	// If we are here, then we should be okay
	return nil
}

func GetBeKindConfig(cfg *rest.Config, ctx context.Context, ns string, name string) ([]byte, error) {
	// Create Kubernetes cilent
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	// Get the secret
	secret, err := client.CoreV1().Secrets(ns).Get(ctx, name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// Get the data
	data, ok := secret.Data["config.yaml"]
	if !ok {
		return nil, errors.New("config.yaml not found in secret")
	}

	// If we are here, then we should be okay
	return data, nil
}

func getPostInstallBytes(m string) ([]byte, error) {
	// Set up []byte to hold the data
	var d []byte

	// Check to see if local file or from web
	switch {
	case strings.HasPrefix(m, "http://"), strings.HasPrefix(m, "https://"):
		data, err := DownloadFileString(m)
		if err != nil {
			return nil, err
		}
		d = []byte(data)
	case strings.HasPrefix(m, "file://"):
		// for a localfile, use the http package to get the file
		t := &http.Transport{}
		t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
		c := &http.Client{Transport: t}
		res, err := c.Get(m)

		if err != nil {
			return nil, err
		}

		defer res.Body.Close()

		d, err = io.ReadAll(res.Body)

		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("only http://, https://, and file:// are supported")
	}

	// Check to see if we even have data
	if len(d) == 0 {
		return nil, errors.New("no data found")
	}

	// If we are here, then we should be okay
	return d, nil
}
