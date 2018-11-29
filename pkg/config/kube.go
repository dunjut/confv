package config

import (
	"fmt"
	"net"

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
	var (
		m    = parseParamString(o.Template)
		name = m["name"]
		key  = m["key"]
		ns   = o.PodNamespace
	)
	cm, err := kube.CoreV1().ConfigMaps(ns).Get(name, meta_v1.GetOptions{})
	if err != nil {
		return "", err
	}
	tpl, ok := cm.Data[key]
	if !ok {
		return "", fmt.Errorf("cannot find config template %s in %s/%s", key, ns, name)
	}
	return tpl, nil
}

// getConfigTemplate retrieves config values from kubernetes configmap
func getConfigValues(kube *kubernetes.Clientset, o *Options) (map[interface{}]interface{}, error) {
	var (
		m            = parseParamString(o.Values)
		name         = m["name"]
		identifiedBy = m["identifiedBy"]
		ns           = o.PodNamespace
	)
	cm, err := kube.CoreV1().ConfigMaps(ns).Get(name, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	pod, err := kube.CoreV1().Pods(o.PodNamespace).Get(o.PodName, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var key string
	switch identifiedBy {
	case "hostIP":
		key = pod.Status.HostIP
	case "nodeName":
		key = pod.Spec.NodeName
	case "podName":
		key = pod.Name
	}

	valuesYaml, ok := cm.Data[key]
	if !ok {
		return nil, fmt.Errorf("cannot find config values for %s(%s) in %s/%s", key, identifiedBy, ns, name)
	}
	valuesMap := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(valuesYaml), &valuesMap)
	return valuesMap, err
}
