package utils

import "sigs.k8s.io/controller-runtime/pkg/client"

// NamespacedName takes a k8s object and returns its namespaced name.
func NamespacedName(obj client.Object) string {
	return client.ObjectKey{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}.String()
}
