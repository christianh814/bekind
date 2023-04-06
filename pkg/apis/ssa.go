/*
Copyright © 2022 Christian Hernandez christian@chernand.io

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package apis

import (
	"context"
	"encoding/json"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

// ApplyResource applies the resource specified in the given YAML using server-side apply
func ApplyResource(ctx context.Context, cfg *rest.Config, resourceYAML []byte) error {
	obj, gvk, err := decodeYAML(resourceYAML)
	if err != nil {
		return err
	}

	mapper, err := getRESTMapper(cfg)
	if err != nil {
		return err
	}

	resourceInterface, err := getResourceInterface(cfg, mapper, gvk, obj)
	if err != nil {
		return err
	}

	return applyResource(ctx, resourceInterface, obj)
}

func decodeYAML(resourceYAML []byte) (*unstructured.Unstructured, *v1.GroupVersionKind, error) {
	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode(resourceYAML, nil, obj)
	if err != nil {
		return nil, nil, err
	}
	return obj, gvk, nil
}

func getRESTMapper(cfg *rest.Config) (meta.RESTMapper, error) {
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc)), nil
}

func getResourceInterface(cfg *rest.Config, mapper meta.RESTMapper, gvk *v1.GroupVersionKind, obj *unstructured.Unstructured) (dynamic.ResourceInterface, error) {
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}

	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		return dyn.Resource(mapping.Resource).Namespace(obj.GetNamespace()), nil
	}
	return dyn.Resource(mapping.Resource), nil
}

func applyResource(ctx context.Context, resourceInterface dynamic.ResourceInterface, obj *unstructured.Unstructured) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	_, err = resourceInterface.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, v1.PatchOptions{
		FieldManager: "bekind",
	})
	return err
}
