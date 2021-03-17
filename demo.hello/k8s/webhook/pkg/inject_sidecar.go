package pkg

import (
	"encoding/json"
	"strings"

	"github.com/golang/glog"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	admissionWebhookAnnotationInjectKey = "sidecar-injector-webhook.test.com/inject"

	podContainerJSONPath     = "/spec/containers"
	podInitContainerJSONPath = "/spec/initContainers"
)

// (https://github.com/kubernetes/kubernetes/issues/57982)
func applyDefaultsWorkaround(sidecarCfg *Config) {
	defaulter.Default(&corev1.Pod{
		Spec: corev1.PodSpec{
			InitContainers: sidecarCfg.InitContainers,
			Containers:     sidecarCfg.Containers,
			Volumes:        sidecarCfg.Volumes,
		},
	})
}

// Check whether the target resoured need to be mutated.
func injectRequired(metadata *metav1.ObjectMeta) bool {
	// skip special kubernete system namespaces
	for _, namespace := range ignoredNamespaces {
		if metadata.Namespace == namespace {
			glog.Infof("Skip mutation for %v for it's in special namespace:%v", metadata.Name, metadata.Namespace)
			return false
		}
	}

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	status := annotations[admissionWebhookAnnotationStatusKey]

	// determine whether to perform mutation based on annotation for the target resource
	var required bool
	if strings.ToLower(status) == "injected" {
		required = false
	} else {
		switch strings.ToLower(annotations[admissionWebhookAnnotationInjectKey]) {
		case "y", "yes", "true", "on":
			required = true
		default:
			required = false
		}
	}

	glog.Infof("Mutation policy for %v/%v: status: %q required: %v", metadata.Namespace, metadata.Name, status, required)
	return required
}

/*
Patch
*/

func addContainer(target, added []corev1.Container, basePath string) (patch []patchOperation) {
	first := len(target) == 0
	var value interface{}
	for _, add := range added {
		value = add
		path := basePath + "/-"
		if first {
			first = false
			path = basePath
			value = []corev1.Container{add}
		}
		patch = append(patch, patchOperation{
			Op:    "add",
			Path:  path,
			Value: value,
		})
	}

	return patch
}

func addVolume(target, added []corev1.Volume) (patch []patchOperation) {
	const basePath = "/spec/volumes"

	first := len(target) == 0
	var value interface{}
	for _, add := range added {
		value = add
		path := basePath + "/-"
		if first {
			first = false
			path = basePath
			value = []corev1.Volume{add}
		}
		patch = append(patch, patchOperation{
			Op:    "add",
			Path:  path,
			Value: value,
		})
	}

	return patch
}

// create mutation patch for resoures from config "sidecar-configmap.yaml".
func createInjectPatch(pod *corev1.Pod, sidecarConfig *Config, annotations map[string]string) ([]byte, error) {
	var patch []patchOperation
	patch = append(patch, addContainer(pod.Spec.InitContainers, sidecarConfig.InitContainers, podInitContainerJSONPath)...)
	patch = append(patch, addContainer(pod.Spec.Containers, sidecarConfig.Containers, podContainerJSONPath)...)
	patch = append(patch, addVolume(pod.Spec.Volumes, sidecarConfig.Volumes)...)
	patch = append(patch, updateAnnotation(pod.Annotations, annotations)...)

	return json.Marshal(patch)
}

/*
Main
*/

func (whsvr *WebhookServer) injectSidecar(ar *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	req := ar.Request
	var pod corev1.Pod
	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		glog.Errorf("Could not unmarshal raw object: %v", err)
		return &admissionv1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	glog.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v (%v) UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, pod.Name, req.UID, req.Operation, req.UserInfo)

	if !injectRequired(&pod.ObjectMeta) {
		glog.Infof("Skipping mutation for %s/%s due to policy check", pod.Namespace, pod.Name)
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	// Workaround: https://github.com/kubernetes/kubernetes/issues/57982
	applyDefaultsWorkaround(whsvr.sidecarCfg)
	annotations := map[string]string{admissionWebhookAnnotationStatusKey: "injected"}
	patchBytes, err := createInjectPatch(&pod, whsvr.sidecarCfg, annotations)
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
