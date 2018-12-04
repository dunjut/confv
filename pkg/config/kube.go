package config

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"gopkg.in/yaml.v2"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	kubeletProxyUnixSocket = "/var/run/confv.sock"
)

// buildKubeClient returns a kubernetes client which talks
// to api server via an unix socket (kubectl proxy).
func buildKubeClient() (*kubernetes.Clientset, error) {
	c := &rest.Config{
		Host: "http://localhost",
		Dial: func(network, addr string) (net.Conn, error) {
			return net.Dial("unix", kubeletProxyUnixSocket)
		},
	}
	return kubernetes.NewForConfig(c)
}

// getConfigTemplate retrieves config template from kubernetes configmap
func getConfigTemplate(kube *kubernetes.Clientset, o *Options) (string, error) {
	if o.Template == "" {
		return "", errors.New("options.template must be specified")
	}
	var (
		ns        = o.PodNamespace
		name, key string
	)
	if pair := strings.SplitN(o.Template, "/", 2); len(pair) == 1 {
		name = o.Template
		// key is omitted by user, expect configmap has only one element
	} else {
		name = pair[0]
		key = pair[1]
	}

	// retrieve template configmap
	cm, err := kube.CoreV1().ConfigMaps(ns).Get(name, meta_v1.GetOptions{})
	if err != nil {
		return "", err
	}

	// if key is set, return its corresponding data
	if key != "" {
		tpl, ok := cm.Data[key]
		if !ok {
			return "", fmt.Errorf("options.template %s not found", o.Template)
		}
		return tpl, nil
	}

	// key is not set, expect configmap has only one element
	if len(cm.Data) != 1 {
		return "", errors.New("template configmap must have only one element when key is not specified")
	}
	for _, tpl := range cm.Data {
		return tpl, nil
	}

	// should not reach here
	return "", errors.New("unexpected!!")
}

// getConfigTemplate retrieves config values from kubernetes configmap
func getConfigValues(kube *kubernetes.Clientset, o *Options) (map[interface{}]interface{}, error) {
	if o.Values == "" || o.IdentifiedBy == "" {
		return nil, errors.New("options.values and options.identifiedBy must be specified")
	}
	var (
		ns   = o.PodNamespace
		name = o.Values
		id   = o.IdentifiedBy
		key  string
	)

	// retrieve values configmap
	cm, err := kube.CoreV1().ConfigMaps(ns).Get(name, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// determin configmap key according to options.identifiedBy
	pod, err := kube.CoreV1().Pods(ns).Get(o.PodName, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	switch id {
	case "hostIP":
		key = pod.Status.HostIP
	case "nodeName":
		key = pod.Spec.NodeName
	case "podName":
		key = pod.Name
	}

	valuesYaml, ok := cm.Data[key]
	if !ok {
		return nil, fmt.Errorf("cannot find config values for %s(%s) in %s", key, id, name)
	}
	valuesMap := make(map[interface{}]interface{})
	if err = yaml.Unmarshal([]byte(valuesYaml), &valuesMap); err != nil {
		return nil, err
	}

	configValues := make(map[interface{}]interface{})
	configValues["values"] = valuesMap
	if o.SharedSecret == "" {
		return configValues, nil
	}

	// inject sharedSecret into configValues
	secret, err := kube.CoreV1().Secrets(ns).Get(o.SharedSecret, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	secretData := make(map[string]string)
	for k, v := range secret.Data {
		secretData[k] = string(v)
	}
	configValues["sharedSecret"] = secretData
	return configValues, nil
}
