package k8s

import (
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	config  *rest.Config
	Dynamic *dynamic.DynamicClient
}

func NewClient(k8sCtx string) (*Client, error) {
	homeDir, _ := os.UserHomeDir()
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: filepath.Join(homeDir, ".kube", "config")},
		&clientcmd.ConfigOverrides{CurrentContext: k8sCtx}).ClientConfig()
	if err != nil {
		return nil, err
	}
	dyn, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Client{
		config:  config,
		Dynamic: dyn,
	}, nil
}

func (c *Client) ListResources() ([]*metav1.APIResourceList, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(c.config)
	if err != nil {
		return nil, err
	}
	return discoveryClient.ServerPreferredResources()
}
