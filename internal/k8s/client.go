package k8s

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	c *kubernetes.Clientset
}

func NewClient(k8sCtx string) (*Client, error) {
	config, err := k8sConfig(k8sCtx)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Client{
		c: clientset,
	}, nil
}

func k8sConfig(k8sCtx string) (*rest.Config, error) {
	homeDir, _ := os.UserHomeDir()
	kubeconfigPath := filepath.Join(homeDir, ".kube", "config")
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{CurrentContext: k8sCtx}).ClientConfig()
}
