package pkg

import (
	"encoding/json"
	"strings"

	"github.com/golang/glog"
	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	admissionWebhookAnnotationMutateKey = "admission-webhook-app.test.com/mutate"
	admissionWebhookAnnotationStatusKey = "admission-webhook-app.test.com/status"

	na = "not_available"
)

var (
	addLabels = map[string]string{
		nameLabel:    na,
		versionLabel: na,
	}
)

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// mutationRequired is mutate op required on request.
func mutationRequired(metadata *metav1.ObjectMeta) bool {
	required := admissionRequired(ignoredNamespaces, admissionWebhookAnnotationMutateKey, metadata)

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	status := annotations[admissionWebhookAnnotationStatusKey]
	if strings.ToLower(status) == "mutated" {
		required = false
	}

	glog.Infof("Mutation policy for %v/%v: required:%v", metadata.Namespace, metadata.Name, required)
	return required
}

/*
Patch
*/

func updateAnnotation(target map[string]string, added map[string]string) (patch []patchOperation) {
	if target == nil {
		target = map[string]string{}
	}

	for key, value := range added {
		if _, ok := target[key]; !ok {
			patch = append(patch, patchOperation{
				Op:   "add",
				Path: "/metadata/annotations",
				Value: map[string]string{
					key: value,
				},
			})
		} else {
			patch = append(patch, patchOperation{
				Op:    "replace",
				Path:  "/metadata/annotations/" + key,
				Value: value,
			})
		}
	}
	return patch
}

func updateLabels(target map[string]string, added map[string]string) (patch []patchOperation) {
	var values map[string]string
	if target == nil {
		values = added
	} else {
		values = make(map[string]string)
		for key, value := range added {
			if target[key] == "" {
				values[key] = value
			}
		}
	}

	patch = append(patch, patchOperation{
		Op:    "add",
		Path:  "/metadata/labels",
		Value: values,
	})
	return patch
}

func createPatch(availableAnnotations map[string]string, annotations map[string]string, availableLabels map[string]string, labels map[string]string) ([]byte, error) {
	patch := updateAnnotation(availableAnnotations, annotations)
	patch = append(patch, updateLabels(availableLabels, labels)...)
	return json.Marshal(patch)
}

/*
Main
*/

func (whsvr *WebhookServer) mutate(ar *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	var (
		availableLabels, availableAnnotations map[string]string
		objectMeta                            *metav1.ObjectMeta
		resourceNamespace, resourceName       string
	)

	req := ar.Request
	glog.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v (%v) UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, resourceName, req.UID, req.Operation, req.UserInfo)

	switch req.Kind.Kind {
	case "Deployment":
		var deployment appsv1.Deployment
		if err := json.Unmarshal(req.Object.Raw, &deployment); err != nil {
			glog.Errorf("Could not unmarshal raw object: %v", err)
			return &admissionv1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		resourceName, resourceNamespace, objectMeta = deployment.Name, deployment.Namespace, &deployment.ObjectMeta
		availableLabels = deployment.Labels
	case "Service":
		var service corev1.Service
		if err := json.Unmarshal(req.Object.Raw, &service); err != nil {
			glog.Errorf("Could not unmarshal raw object: %v", err)
			return &admissionv1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		resourceName, resourceNamespace, objectMeta = service.Name, service.Namespace, &service.ObjectMeta
		availableLabels = service.Labels
	}

	if !mutationRequired(objectMeta) {
		glog.Infof("Skipping mutation for %s/%s due to policy check", resourceNamespace, resourceName)
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	annotations := map[string]string{admissionWebhookAnnotationStatusKey: "mutated"}
	patchBytes, err := createPatch(availableAnnotations, annotations, availableLabels, addLabels)
	if err != nil {
		return &admissionv1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	glog.Infof("AdmissionResponse: patch=%v\n", string(patchBytes))
	return &admissionv1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *admissionv1.PatchType {
			pt := admissionv1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}
