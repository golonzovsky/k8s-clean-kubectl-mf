package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	config    *rest.Config
	Dynamic   *dynamic.DynamicClient
	Discovery *discovery.DiscoveryClient
}

func NewClient(k8sCtx string) (*Client, error) {
	config, err := localOrInClusterConfig(k8sCtx)
	if err != nil {
		return nil, err
	}
	dyn, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	disc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Client{
		Dynamic:   dyn,
		Discovery: disc,
	}, nil
}

func localOrInClusterConfig(k8sCtx string) (*rest.Config, error) {
	homeDir, _ := os.UserHomeDir()
	kubeConfig := filepath.Join(homeDir, ".kube", "config")

	if os.Getenv("KUBECONFIG") != "" {
		kubeConfig = os.Getenv("KUBECONFIG")
	}

	_, err := os.Stat(kubeConfig)
	if err == nil {
		return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfig},
			&clientcmd.ConfigOverrides{CurrentContext: k8sCtx}).ClientConfig()
	}
	if os.IsNotExist(err) {
		if k8sCtx != "" {
			return nil, fmt.Errorf("k8s context override flag is only supported for local client")
		}
		return rest.InClusterConfig()
	}
	return nil, err
}

func (c *Client) ListResources() ([]*metav1.APIResourceList, error) {
	return c.Discovery.ServerPreferredResources()
}
