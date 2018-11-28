package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"text/template"
)

// Options is user specified volume options in the pod spec.
type Options struct {
	PodName        string `json:"kubernetes.io/pod.name"`
	PodNamespace   string `json:"kubernetes.io/pod.namespace"`
	Template       string `json:"template"`
	Values         string `json:"values"`
	IdentifiedBy   string `json:"identifiedBy"`
	TargetFileName string `json:"targetFileName"`
}

func (o *Options) Validate() error {
	if o.PodName == "" || o.PodNamespace == "" {
		return errors.New("pod name and namespace must be set")
	}

	var m map[string]string
	m = parseParamString(o.Template)
	if m["name"] == "" || m["key"] == "" {
		return errors.New("options.template must have configmap name and key")
	}

	m = parseParamString(o.Values)
	if m["name"] == "" {
		return errors.New("options.values must have configmap name")
	}

	switch o.IdentifiedBy {
	case "hostIP", "nodeName", "podName":
	default:
		return errors.New("options.identifiedBy must be one of hostIP|nodeName|podName")
	}

	if o.TargetFileName == "" {
		return errors.New("options.targetFileName must be set")
	}
	return nil
}

// DecodeOptions decodes option values from raw json options.
func DecodeOptions(rawOptions string) (*Options, error) {
	o := new(Options)
	if err := json.Unmarshal([]byte(rawOptions), o); err != nil {
		return nil, err
	}

	return o, o.Validate()
}

// RenderConfig returns raw bytes of rendered config.
func RenderConfig(o *Options) ([]byte, error) {
	// build kubernetes client which talks to apiserver via kubelet proxy (unix socket)
	kube, err := buildKubeClient()
	if err != nil {
		return nil, fmt.Errorf("buildKubeClient: %v", err)
	}

	// retrieve config template
	tpl, err := getConfigTemplate(kube, o)
	if err != nil {
		return nil, fmt.Errorf("getConfigTemplate: %v", err)
	}

	// retrieve config values
	vals, err := getConfigValues(kube, o)
	if err != nil {
		return nil, fmt.Errorf("getConfigValues: %v", err)
	}

	// render template with values
	t, err := template.New("gotpl").Parse(tpl)
	if err != nil {
		return nil, fmt.Errorf("parse template: %v", err)
	}
	renderedCfg := &bytes.Buffer{}
	if err = t.Execute(renderedCfg, vals); err != nil {
		return nil, fmt.Errorf("render: %v", err)
	}

	return renderedCfg.Bytes(), nil
}

func parseParamString(s string) map[string]string {
	m := make(map[string]string)
	for _, kv := range strings.Split(s, ",") {
		if res := strings.Split(kv, "="); len(res) == 2 {
			k := strings.TrimSpace(res[0])
			v := strings.TrimSpace(res[1])
			m[k] = v
		}
	}
	return m
}
