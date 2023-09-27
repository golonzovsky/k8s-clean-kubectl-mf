package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	log.SetLevel(log.DebugLevel)
	errorLogger := log.New(os.Stderr)

	if err := doRun(context.Background()); err != nil {
		if !errors.Is(err, context.Canceled) {
			errorLogger.Error(err.Error())
		}
	}
}

func doRun(ctx context.Context) error {
	config, err := k8sConfig("dev-cookie")
	if err != nil {
		return err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	d, err := client.AppsV1().Deployments("platform").Get(ctx, "sre-api-primary", metav1.GetOptions{})
	if err != nil {
		return err
	}
	filteredManagedFields := make([]metav1.ManagedFieldsEntry, 0)
	for _, f := range d.ObjectMeta.ManagedFields {
		log.Infof("manager: %s, Operation: %s, Time: %s, APIVersion: %s, FieldsType: %s: %s",
			f.Manager, f.Operation, f.Time.Format("2023-09-27 12:41:01"), f.APIVersion, f.FieldsType, f.FieldsV1)

		if f.Manager == "kubectl-edit" {
			log.Warnf("found managed field from kubectl-edit: %s", f)
		} else {
			filteredManagedFields = append(filteredManagedFields, f)
		}
	}
	d.ObjectMeta.ManagedFields = filteredManagedFields

	_, err = client.AppsV1().Deployments("platform").Update(context.TODO(), d, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func k8sConfig(k8sCtx string) (*rest.Config, error) {
	homeDir, _ := os.UserHomeDir()
	kubeconfigPath := filepath.Join(homeDir, ".kube", "config")
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{CurrentContext: k8sCtx}).ClientConfig()
}
