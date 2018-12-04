package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
)

// Options is user specified volume options in the pod spec.
type Options struct {
	PodName      string `json:"kubernetes.io/pod.name"`
	PodNamespace string `json:"kubernetes.io/pod.namespace"`
	Template     string `json:"template"`
	Values       string `json:"values"`
	IdentifiedBy string `json:"identifiedBy"`
	SharedSecret string `json:"sharedSecret"`
	Target       string `json:"target"`
}

// DecodeOptions decodes option values from raw json options.
func DecodeOptions(rawOptions string) (*Options, error) {
	o := new(Options)
	return o, json.Unmarshal([]byte(rawOptions), o)
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
