package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	k8sutils "demo.hello/k8s/client/pkg"
	webshell "demo.hello/k8s/webshell/pkg"
	"demo.hello/utils"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	defaultPath = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	kubeConfig  = flag.String("kubeconfig", defaultPath, "abs path to the kubeconfig file")
	addr        = flag.String("addr", ":8090", "http service address")
	cmd         = []string{"/bin/sh"}
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/query/ns", getAllNamespaces)
	router.HandleFunc("/query/pods", getAllPodsByNamespace)

	router.HandleFunc("/terminal", serveTerminal)
	router.HandleFunc("/ws/{namespace}/{pod}/{container_name}/webshell", serveWs)

	log.Printf("http server (websocket) is started at :%s...", *addr)
	log.Fatal(http.ListenAndServe(*addr, router))
}

//
// K8s resources query api (json)
// Namespaces: curl -v "http://localhost:8090/query/ns" | jq .
// Pods: curl -v "http://localhost:8090/query/pods?ns=k8s-test" | jq .
//

type respJSONData struct {
	Meta json.RawMessage `json:"meta,omitempty"`
	Data json.RawMessage `json:"data"`
}

type respErrorMsg struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type respK8SResources struct {
	Namespaces []string `json:"namespaces,omitempty"`
	Pods       []string `json:"pods,omitempty"`
}

func getAllNamespaces(w http.ResponseWriter, r *http.Request) {
	resource, err := createK8sResourceClient(w, *kubeConfig)
	if err != nil {
		writeJSONRespWithStatus(w, http.StatusInternalServerError, err.Error())
	}

	namespaces, err := resource.GetAllNamespace(context.Background())
	if err != nil {
		err = fmt.Errorf("Get k8s all namespaces error: %v", err)
		writeJSONRespWithStatus(w, http.StatusInternalServerError, err.Error())
	}

	NsNames := make([]string, 0, len(namespaces))
	for _, ns := range namespaces {
		NsNames = append(NsNames, ns.GetName())
	}
	b, err := json.Marshal(respK8SResources{Namespaces: NsNames})
	if err != nil {
		err = fmt.Errorf("json marshal error: %v", err)
		writeJSONRespWithStatus(w, http.StatusInternalServerError, err.Error())
	}
	writeOkJSONResp(w, string(b))
}

func getAllPodsByNamespace(w http.ResponseWriter, r *http.Request) {
	namespace := "default"
	values := r.URL.Query()
	if val, ok := values["ns"]; ok {
		namespace = val[0]
	} else {
		log.Printf("Use default namespace [%s] to query pods\n", namespace)
	}

	resource, err := createK8sResourceClient(w, *kubeConfig)
	if err != nil {
		writeJSONRespWithStatus(w, http.StatusInternalServerError, err.Error())
	}
	pods, err := resource.GetPodsByNamespace(context.Background(), namespace)
	if err != nil {
		err = fmt.Errorf("Get k8s pods in namespace [%s] error: %v", namespace, err)
		writeJSONRespWithStatus(w, http.StatusInternalServerError, err.Error())
	}

	podNames := make([]string, 0, len(pods))
	for _, pod := range pods {
		podNames = append(podNames, pod.GetName())
	}
	b, err := json.Marshal(respK8SResources{Pods: podNames})
	if err != nil {
		err = fmt.Errorf("json marshal error: %v", err)
		writeJSONRespWithStatus(w, http.StatusInternalServerError, err.Error())
	}
	writeOkJSONResp(w, string(b))
}

func createK8sResourceClient(w http.ResponseWriter, kubeConfig string) (*k8sutils.Resource, error) {
	client, err := k8sutils.CreateK8sClientLocal(kubeConfig)
	if err != nil {
		err = fmt.Errorf("Init k8s client error: %v", err)
		return nil, err

	}
	return k8sutils.NewResource(client), nil
}

//
// K8s terminal session
//

func internalError(conn *websocket.Conn, msg string, err error) {
	err = fmt.Errorf("Internal server error.\nmessage: %s, error: %v", msg, err)
	conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
}

func serveTerminal(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "./static/terminal.html")
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	namespace := pathParams["namespace"]
	pod := pathParams["pod"]
	containerName := pathParams["container_name"]
	log.Printf("ws request: exec pod:%s, container:%s, namespace:%s", pod, containerName, namespace)

	term, err := webshell.NewTerminalSession(w, r, nil)
	if err != nil {
		log.Printf("get terminal session failed: %v", err)
		return
	}
	defer func() {
		log.Println("close session")
		term.Close()
	}()

	resource, err := createK8sResourceClient(w, *kubeConfig)
	if err != nil {
		log.Printf("create k8s resource client failed: %v", err)
	}

	if containerName != "null" {
		ok, err := resource.IsPodExec(r.Context(), namespace, pod, containerName)
		if !ok {
			log.Printf("check pod failed: pod:%s, container:%s, namespace:%s\n", pod, containerName, namespace)
			if err != nil {
				msg := fmt.Sprintf("validate pod error: %v", err)
				log.Println(msg)
				term.Write([]byte(msg))
				term.Done()
			}
			return
		}
	} else {
		pod, err := resource.GetPod(r.Context(), namespace, pod)
		if err != nil {
			log.Printf("get pod failed: pod:%s, namespace:%s\n", pod, namespace)
			return
		}
		containerName = pod.Spec.Containers[0].Name
	}

	client := resource.GetClient()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err != nil {
		log.Printf("create k8s config failed: %v\n", err)
	}
	if err := webshell.ExecPod(client, config, cmd, term, namespace, pod, containerName); err != nil {
		msg := fmt.Sprintf("exec pod error: %v", err)
		log.Println(msg)
		term.Write([]byte(msg))
		term.Done()
	}
}

//
// http io
//

func writeOkJSONResp(w http.ResponseWriter, data string) {
	writeJSONRespWithStatus(w, http.StatusOK, data)
}

func writeJSONRespWithStatus(w http.ResponseWriter, retCode int, data string) {
	utils.AddCorsHeadersForOptions(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(retCode)

	if err := json.NewEncoder(w).Encode(respJSONData{Data: json.RawMessage(data)}); err != nil {
		log.Println("Write json encoded response error:", err.Error())
	}
}
