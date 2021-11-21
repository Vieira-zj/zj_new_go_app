# K8S Admission Webhook

> k8s admission webhook demo for init container and sidecar injector.
> 

Refer:

- <https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#response>
- <https://github.com/morvencao/kube-mutating-webhook-tutorial>

## K8S env build

1. Check k8s APIs are ok for webhook.

```sh
# check cluster
minikube status

kubectl api-versions | grep admission
# admissionregistration.k8s.io/v1
# admissionregistration.k8s.io/v1beta1

kubectl get pods kube-apiserver-minikube -n kube-system -o yaml | grep enable
# --enable-admission-plugins=NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook
```

2. Create security for webhook https api.

```sh
./deploy/webhook-create-signed-cert.sh

kubectl get secret admission-webhook-app-certs -n k8s-test
# admission-webhook-app-certs   Opaque   2      49s
```

3. Compile server bin, build and push image.

```sh
eval $(minikube -p minikube docker-env)
./run.sh build
./run.sh push

# check image
docker images
docker run --name webhook -d --rm zhengjin.k8s.test/admission-webhook-app:v1.0.0
```

4. Create webhook deployment and service in k8s.

```sh
kubectl create -f deploy/webhook-deployment.yaml
kubectl create -f deploy/webhook-service.yaml

# check deploy resources
kubectl get all -n k8s-test

# check webhook service
kubectl run debug -it --rm --restart=Never --namespace=k8s-test --image=busybox:1.30 sh
wget "https://admission-webhook-app-svc/validate" --no-check-certificate
# server logs
kubectl logs pod/admission-webhook-app-deployment-5d44f5db7b-ghmvh -n k8s-test
# I0104 07:53:04.658182       1 main.go:63] Server started
# E0104 07:56:08.807973       1 webhook_server.go:51] empty body
```

## Validate webhook for labels

1. Add label for test namespace which is used for validation and mutating webhook.

```sh
kubectl label namespace k8s-test admission-webhook-app=enabled
# check
kubectl get namespace k8s-test -o yaml
```

2. Set token and deploy `ValidatingWebhookConfiguration`, which binds external webhook `/validate` to validating admission control.

```sh
# 获得 token 替换 validatingwebhook-ca-bundle.yaml ${CA_BUNDLE}
kubectl config view --raw --flatten -o json | grep certificate-authority-data
# create and check bundle
kubectl create -f deploy/validatingwebhook-ca-bundle.yaml

# check
kubectl get validatingwebhookconfiguration
# NAME                                WEBHOOKS   AGE
# labels-validation-webhook-app-cfg   1          7s
```

3. Test validation webhook.

- Case1: without labels, and deploy fail

```sh
kubectl create -f deploy/sleep-deployment.yaml
# Error from server (required label [app.kubernetes.io/name] are not set): 
# error when creating "deploy/sleep-deployment.yaml": 
# admission webhook "required-labels.test.com" denied the request: 
# required label [app.kubernetes.io/name] are not set
```

- Case2: with required labels, and deploy pass

```sh
# add labels in deploy yaml:
# labels:
#   app.kubernetes.io/name: sleep
#   app.kubernetes.io/version: "1.0"

kubectl create -f deploy/sleep-deployment.yaml
# deployment.apps/sleep created
```

- Case3: with validation is set as `false`, and deploy apss 

```sh
# add validate annotation in deploy yaml:
# annotations:
#   admission-webhook-app.test.com/validate: "false"

kubectl create -f deploy/sleep-deployment.yaml
```

## Mutate webhook for labels

1. Set token and deploy `kind: MutatingWebhookConfiguration`, which binds external webhook `/mutate` to mutating admission control.

```sh
# 获得 token 替换 mutatingwebhook-ca-bundle.yaml ${CA_BUNDLE}
kubectl config view --raw --flatten -o json | grep certificate-authority-data

# create and check bundle
kubectl create -f deploy/mutatingwebhook-ca-bundle.yaml
kubectl get mutatingwebhookconfiguration
# NAME                              WEBHOOKS   AGE
# labels-mutating-webhook-app-cfg   1          7s
```

2. Test mutations are applied.

Create sleep deploy with removing labels and validation `false` annotation.

```sh
kubectl create -f deploy/sleep-deployment.yaml

# check added labels and annotation
kubectl get deployment/sleep -n k8s-test -o yaml
# annotations:
#   admission-webhook-example.banzaicloud.com/status: mutated
# labels:
#   app.kubernetes.io/name: not_available
#   app.kubernetes.io/version: not_available
```

> Note: k8s apply mutate webhook first, and then exec validate webhook, so we success create deployment without labels.
>

3. Check logs of admission webhook server.

```sh
kubectl logs pod/admission-webhook-app-deployment-5d44f5db7b-ghmvh -n k8s-test

# without required labels, and validate failed
# I0107 02:02:47.463061] /validate
# I0107 02:02:47.463157] AdmissionReview for Kind=apps/v1, Kind=Deployment, Namespace=k8s-test Name=sleep () UID=c4398009-4ed5-4c5e-9f03-87dadc8ada9e patchOperation=CREATE UserInfo={minikube-user  [system:masters system:authenticated] map[]}
# I0107 02:02:47.465434] Validation policy for k8s-test/sleep: required:true
# I0107 02:02:47.465596] available labels:map[]
# I0107 02:02:47.465668] required labels[app.kubernetes.io/name app.kubernetes.io/version]
# I0107 02:02:47.465961] Ready to write reponse ...

# add required labels
# I0107 02:16:27.919923] /mutate
# I0107 02:16:27.919946] AdmissionReview for Kind=apps/v1, Kind=Deployment, Namespace=k8s-test Name=sleep () UID=1283813a-5783-47d3-8f27-43ceb17e6d27 patchOperation=CREATE UserInfo={minikube-user  [system:masters system:authenticated] map[]}
# I0107 02:16:27.920113] Mutation policy for k8s-test/sleep: required:true
# I0107 02:16:27.920170] AdmissionResponse: patch=[{"op":"add","path":"/metadata/annotations","value":{"admission-webhook-app.test.com/status":"mutated"}},{"op":"add","path":"/metadata/labels","value":{"app.kubernetes.io/name":"not_available","app.kubernetes.io/version":"not_available"}}]
# I0107 02:16:27.920198] Ready to write reponse ...

# validate pass
# I0107 02:16:27.922245] /validate
# I0107 02:16:27.922273] AdmissionReview for Kind=apps/v1, Kind=Deployment, Namespace=k8s-test Name=sleep () UID=b27d2024-992d-4e43-8c89-2ed0116b4205 patchOperation=CREATE UserInfo={minikube-user  [system:masters system:authenticated] map[]}
# I0107 02:16:27.922441] Validation policy for k8s-test/sleep: required:true
# I0107 02:16:27.922459] available labels:map[app.kubernetes.io/name:not_available app.kubernetes.io/version:not_available]
# I0107 02:16:27.922491] required labels[app.kubernetes.io/name app.kubernetes.io/version]
# I0107 02:16:27.922503] Ready to write reponse ...
```

## Inject sidecar (busybox)

1. Deploy configmap which include webhook inject sidecar configs.

```sh
kubectl create -f deploy/injector-configmap.yaml
kubectl get cm -n k8s-test
```

2. Create webhook deployment (which mounts sidecar configmap) and service.

```sh
kubectl create -f deploy/webhook-deployment.yaml
kubectl create -f deploy/webhook-service.yaml
```

3. Create injector bundle which binds external webhook `/inject` to mutating admission control.

```sh
kubectl create -f deploy/injectorwebhook-ca-bundle.yaml
```

4. Apply sleep deployment with pod annotations `sidecar-injector-webhook.test.com/inject: "true"`

```sh
kubectl create -f deploy/sleep-deployment.yaml

# check deploy pod
kubectl get pod -n k8s-test
# NAME                                                READY   STATUS    RESTARTS   AGE
# sleep-5659658d59-qzjlm                              2/2     Running   0          60s

kubectl get pod/sleep-5659658d59-qzjlm -n k8s-test -o yaml
# annotations:
#   admission-webhook-app.test.com/status: injected
# containers:
# name: sidecar-busybox

kubectl logs pod/sleep-5659658d59-qzjlm -c sidecar-busybox -n k8s-test
```

5. Check admission webhook server logs.

```sh
kubectl logs pod/admission-webhook-app-deployment-659c9fc46b-7wcr7 -n k8s-test

# I0109 07:38:13.286192] /inject
# I0109 07:38:13.288288] AdmissionReview for Kind=/v1, Kind=Pod, Namespace=k8s-test Name= () UID=f730e36c-26ed-46c5-a714-81035170c53f patchOperation=CREATE UserInfo={system:serviceaccount:kube-system:replicaset-controller 9073ae99-d884-4f97-bee8-b23051037346 [system:serviceaccounts system:serviceaccounts:kube-system system:authenticated] map[]}
# I0109 07:38:13.288350] Mutation policy for /: status: "" required: true
# I0109 07:38:13.288421] AdmissionResponse: patch=[{"op":"add","path":"/spec/containers/-","value":{"name":"sidecar-busybox","image":"busybox:1.30","command":["sh","-c","while true; do echo $(date +'%Y-%m-%d_%H:%M:%S') 'busybox is running ...'; sleep 5; done;"],"resources":{}}},{"op":"add","path":"/metadata/annotations","value":{"admission-webhook-app.test.com/status":"injected"}}]
# I0109 07:38:13.288594] Ready to write reponse: {"kind":"AdmissionReview","apiVersion":"admission.k8s.io/v1","response":{"uid":"f730e36c-26ed-46c5-a714-81035170c53f","allowed":true,"patch":"W3sib3AiOiJhZGQiLCJwYXRoIjoiL3NwZWMvY29udGFpbmVycy8tIiwidmFsdWUiOnsibmFtZSI6InNpZGVjYXItYnVzeWJveCIsImltYWdlIjoiYnVzeWJveDoxLjMwIiwiY29tbWFuZCI6WyJzaCIsIi1jIiwid2hpbGUgdHJ1ZTsgZG8gZWNobyAkKGRhdGUgKyclWS0lbS0lZF8lSDolTTolUycpICdidXN5Ym94IGlzIHJ1bm5pbmcgLi4uJzsgc2xlZXAgNTsgZG9uZTsiXSwicmVzb3VyY2VzIjp7fX19LHsib3AiOiJhZGQiLCJwYXRoIjoiL21ldGFkYXRhL2Fubm90YXRpb25zIiwidmFsdWUiOnsiYWRtaXNzaW9uLXdlYmhvb2stYXBwLnRlc3QuY29tL3N0YXR1cyI6ImluamVjdGVkIn19XQ==","patchType":"JSONPatch"}}
```

## Inject init container

Workflow:

```sh
# 1: prepare webhook image
./run.sh build
./run.sh push

# 2: deploy admission webhook server
# deploy configmap used by webhook
kubectl create -f deploy/injector-configmap.yaml
# deploy webhook
kubectl create -f deploy/webhook-deployment.yaml
# deploy bundle
kubectl create -f deploy/injectorwebhook-ca-bundle.yaml

# 3: apply a deployment and check
kubectl create -f deploy/sleep-deployment.yaml

# check pod
kubectl get pods -n k8s-test
# NAME                                                READY   STATUS            RESTARTS   AGE
# sleep-5659658d59-2mdxk                              0/2     PodInitializing   0          6s

# check pod events
kubectl describe pod/sleep-5659658d59-2mdxk -n k8s-test
# Normal  Scheduled  19h   default-scheduler  Successfully assigned k8s-test/sleep-5659658d59-2mdxk to minikube
# Normal  Pulled     19h   kubelet            Container image "busybox:1.30" already present on machine
# Normal  Created    19h   kubelet            Created container init-busybox
# Normal  Started    19h   kubelet            Started container init-busybox
# Normal  Pulled     19h   kubelet            Container image "tutum/curl" already present on machine
# Normal  Created    19h   kubelet            Created container sleep
# Normal  Started    19h   kubelet            Started container sleep
# Normal  Pulled     19h   kubelet            Container image "busybox:1.30" already present on machine
# Normal  Created    19h   kubelet            Created container sidecar-busybox
# Normal  Started    19h   kubelet            Started container sidecar-busybox

# check pod logs
kc logs sleep-5659658d59-2mdxk
# choose one of: [sleep sidecar-busybox] or one of the init containers: [init-busybox]

# check pod definition
kubectl get pod/sleep-5659658d59-2mdxk -n k8s-test -o yaml
# annotations:
#   admission-webhook-app.test.com/status: injected
# initContainers:
#   name: init-busybox
# containers:
#   name: sidecar-busybox

# check webhook server logs
kubectl logs -n k8s-test admission-webhook-app-deployment-659c9fc46b-jlx7n
# I0109 11:24:18.718925] AdmissionReview for Kind=/v1, Kind=Pod, Namespace=k8s-test Name= () UID=cac8ab0c-8ef5-4bc4-aedf-c9b4e1f08c25 patchOperation=CREATE UserInfo={system:serviceaccount:kube-system:replicaset-controller 9073ae99-d884-4f97-bee8-b23051037346 [system:serviceaccounts system:serviceaccounts:kube-system system:authenticated] map[]}
# I0109 11:24:18.719078] Mutation policy for /: status: "" required: true
# I0109 11:24:18.719294] AdmissionResponse: patch=[{"op":"add","path":"/spec/initContainers","value":[{"name":"init-busybox","image":"busybox:1.30","command":["sh","-c","sleep 3; echo \"init process done.\""],"resources":{}}]},{"op":"add","path":"/spec/containers/-","value":{"name":"sidecar-busybox","image":"busybox:1.30","command":["sh","-c","while true; do echo $(date +'%Y-%m-%d_%H:%M:%S') 'busybox is running ...'; sleep 5; done;"],"resources":{}}},{"op":"add","path":"/metadata/annotations","value":{"admission-webhook-app.test.com/status":"injected"}}]

# 4: clear injector resources
kubectl delete -f deploy/sleep-deployment.yaml \
  && kc delete -f deploy/injectorwebhook-ca-bundle.yaml \
  && kc delete -f deploy/webhook-deployment.yaml
```

