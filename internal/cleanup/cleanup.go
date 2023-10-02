package cleanup

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/ricardo/k8s-managed-field-cleanup/internal/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func DoRunCleanup(ctx context.Context, dryRun bool) error {
	client, err := k8s.NewClient("dev-cookie")
	if err != nil {
		return err
	}

	apiResourceList, err := client.ListResources()
	if err != nil {
		return err
	}

	for _, resources := range apiResourceList {
		gv, err := schema.ParseGroupVersion(resources.GroupVersion)
		if err != nil {
			return err
		}
		for _, res := range resources.APIResources {
			gvr := schema.GroupVersionResource{
				Group:    gv.Group,
				Version:  gv.Version,
				Resource: res.Name,
			}

			if !res.Namespaced { // todo process cluster wide as well
				log.Debug("Skipping cluster-wide resource", "resource", gvr.String())
				continue
			}
			log.Debug("Checking namespaced", "resource", gvr.String())

			resourceList, err := client.Dynamic.Resource(gvr).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return err
				}
				log.Error("Error listing "+gvr.String()+":", "err", err)
				continue
			}

			for _, obj := range resourceList.Items {
				filteredFields := make([]metav1.ManagedFieldsEntry, 0)
				for _, f := range obj.GetManagedFields() {
					if f.Manager == "kubectl-edit" {
						log.Warn("found kubectl-edit managed field",
							"gvr", fmt.Sprintf("%s/%s/%s", gvr.Group, gvr.Version, gvr.Resource),
							"res", obj.GetNamespace()+"/"+obj.GetName(),
							"field", f)
						continue
					}
					filteredFields = append(filteredFields, f)
				}
				if len(filteredFields) == len(obj.GetManagedFields()) {
					continue
				}
				if s, ok := obj.GetLabels()["app.kubernetes.io/managed-by"]; !ok || s != "Helm" && s != "flagger" {
					log.Debug("Skipping non-Helm/Flagger resource with kubectl fields", "resource", gvr.String(), "name", obj.GetName())
					continue
				}
				obj.SetManagedFields(filteredFields)
				options := metav1.UpdateOptions{}
				if dryRun {
					continue
					//options.DryRun = []string{metav1.DryRunAll}
				}
				_, err = client.Dynamic.Resource(gvr).Namespace(obj.GetNamespace()).Update(ctx, &obj, options)
				if err != nil {
					return err
				}
				log.Infof("updated %s/%s/%s: %s", gvr.Group, gvr.Version, gvr.Resource, obj.GetName())
			}
		}
	}
	return nil
}
