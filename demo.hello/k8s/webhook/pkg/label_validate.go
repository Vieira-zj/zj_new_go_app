package pkg

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang/glog"
	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	admissionWebhookAnnotationValidateKey = "admission-webhook-app.test.com/validate"

	nameLabel    = "app.kubernetes.io/name"
	versionLabel = "app.kubernetes.io/version"
)

var (
	ignoredNamespaces = []string{
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	}

	requiredLabels = []string{
		nameLabel,
		versionLabel,
	}
)

func admissionRequired(ignoredList []string, admissionAnnotationKey string, metadata *metav1.ObjectMeta) bool {
	// skip special kubernetes system namespaces
	for _, namespace := range ignoredList {
		if metadata.Namespace == namespace {
			glog.Infof("Skip validation for %v for it's in special namespace:%v", metadata.Name, metadata.Namespace)
			return false
		}
	}

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	// check validate key set, default "required"
	switch strings.ToLower(annotations[admissionAnnotationKey]) {
	case "n", "no", "false", "off":
		return false
	default:
		return true
	}
}

// validationRequired is valdate op required on request.
func validationRequired(metadata *metav1.ObjectMeta) bool {
	required := admissionRequired(ignoredNamespaces, admissionWebhookAnnotationValidateKey, metadata)
	glog.Infof("Validation policy for %v/%v: required:%v", metadata.Namespace, metadata.Name, required)
	return required
}

// check labels for deployments and services.
func (whsvr *WebhookServer) validate(ar *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	var (
		availableLabels                 map[string]string
		objectMeta                      *metav1.ObjectMeta
		resourceNamespace, resourceName string
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

	if !validationRequired(objectMeta) {
		glog.Infof("Skipping validation for %s/%s due to policy check", resourceNamespace, resourceName)
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	allowed := true
	var result *metav1.Status
	glog.Info("available labels:", availableLabels)
	glog.Info("required labels", requiredLabels)

	for _, rlabel := range requiredLabels {
		if _, ok := availableLabels[rlabel]; !ok {
			allowed = false
			result = &metav1.Status{
				Reason: metav1.StatusReason(fmt.Sprintf("required label [%v] are not set", rlabel)),
			}
			break
		}
	}

	return &admissionv1.AdmissionResponse{
		Allowed: allowed,
		Result:  result,
	}
}
