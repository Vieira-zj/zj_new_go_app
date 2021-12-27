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
)

var (
	defaultPath = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	kubeConfig  = flag.String("wskubeconfig", defaultPath, "abs path to the kubeconfig file")
	addr        = flag.String("addr", ":8090", "http service address")
)

func main() {
	go func() {
		// here, should use "/" for file server
		http.Handle("/", http.FileServer(http.Dir("/tmp/test")))
		http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "ok")
		})

		fsAddr := ":8091"
		log.Printf("file server (websocket) is started at %s...\n", fsAddr)
		log.Fatal(http.ListenAndServe(fsAddr, nil))
	}()

	router := mux.NewRouter()
	router.HandleFunc("/query/ns", getAllNamespaces)
	router.HandleFunc("/query/pods", getAllPodsByNamespace)
	router.HandleFunc("/query/containers", getAllContainersByPod)

	router.HandleFunc("/terminal", serveTerminal)
	router.HandleFunc("/ws/{namespace}/{pod}/{container_name}/webshell", serveWs)

	log.Printf("http server (websocket) is started at %s...\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, router))
}

//
// K8s resources query api
//
// Namespaces: curl -v "http://localhost:8090/query/ns" | jq .
// Pods: curl -v "http://localhost:8090/query/pods?ns=k8s-test" | jq .
// Containers: curl -v "http://localhost:8090/query/containers?ns=k8s-test&pod=containers-pod" | jq .
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
	Containers []string `json:"containers,omitempty"`
}

func getAllNamespaces(w http.ResponseWriter, r *http.Request) {
	resource, err := createK8sResourceClient()
	if err != nil {
		err = fmt.Errorf("init k8s client error: %v", err)
		writeJSONRespWithStatus(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}

	nsNames, err := resource.GetAllNamespacesName(context.Background())
	if err != nil {
		err = fmt.Errorf("get k8s all namespaces error: %v", err)
		writeJSONRespWithStatus(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}

	b, err := json.Marshal(respK8SResources{Namespaces: nsNames})
	if err != nil {
		err = fmt.Errorf("json marshal error: %v", err)
		writeJSONRespWithStatus(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}

	log.Println("get cluster all namespaces")
	writeOkJSONResp(w, b)
}

func getAllPodsByNamespace(w http.ResponseWriter, r *http.Request) {
	resource, err := createK8sResourceClient()
	if err != nil {
		err = fmt.Errorf("init k8s client error: %v", err)
		writeJSONRespWithStatus(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}

	namespace := "default"
	values := r.URL.Query()
	if val, ok := values["ns"]; ok {
		namespace = val[0]
	} else {
		log.Printf("use default namespace [%s] to query pods\n", namespace)
	}

	podNames, err := resource.GetPodsNameByNamespace(context.Background(), namespace)
	if err != nil {
		err = fmt.Errorf("get k8s pods in namespace [%s] error: %v", namespace, err)
		writeJSONRespWithStatus(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}

	b, err := json.Marshal(respK8SResources{Pods: podNames})
	if err != nil {
		err = fmt.Errorf("json marshal error: %v", err)
		writeJSONRespWithStatus(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}
	log.Printf("get namespace [%s] all pods\n", namespace)
	writeOkJSONResp(w, b)
}

func getAllContainersByPod(w http.ResponseWriter, r *http.Request) {
	var namespace, podName string
	values := r.URL.Query()
	if val, ok := values["ns"]; ok {
		namespace = val[0]
	} else {
		errResp := respErrorMsg{
			Status:  499,
			Message: "Namespace is not set in query of request url",
		}
		b, err := json.Marshal(errResp)
		if err != nil {
			err = fmt.Errorf("json marshal error: %v", err)
			writeJSONRespWithStatus(w, http.StatusInternalServerError, []byte(err.Error()))
			return
		}
		writeJSONRespWithStatus(w, http.StatusNotAcceptable, b)
		return
	}

	if val, ok := values["pod"]; ok {
		podName = val[0]
	} else {
		errResp := respErrorMsg{
			Status:  499,
			Message: "Pod is not set in query of request url",
		}
		b, err := json.Marshal(errResp)
		if err != nil {
			err = fmt.Errorf("json marshal error: %v", err)
			writeJSONRespWithStatus(w, http.StatusInternalServerError, []byte(err.Error()))
			return
		}
		writeJSONRespWithStatus(w, http.StatusNotAcceptable, b)
		return
	}

	resource, err := createK8sResourceClient()
	if err != nil {
		err = fmt.Errorf("init k8s client error: %v", err)
		writeJSONRespWithStatus(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}

	containers, err := resource.GetPodCotainersName(context.Background(), namespace, podName)
	if err != nil {
		log.Printf("get pod [%s/%s] all containers error: %s\n", namespace, podName, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := respK8SResources{Containers: containers}
	b, err := json.Marshal(resp)
	if err != nil {
		err = fmt.Errorf("json marshal error: %v", err)
		writeJSONRespWithStatus(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}
	log.Printf("get pod [%s/%s] all containers\n", namespace, podName)
	writeOkJSONResp(w, b)
}

func createK8sResourceClient() (*k8sutils.Resource, error) {
	client, err := k8sutils.CreateK8sClientLocal(*kubeConfig)
	if err != nil {
		err = fmt.Errorf("Init k8s client error: %v", err)
		return nil, err
	}
	return k8sutils.NewResource(client), nil
}

//
// K8s terminal session by websocket
//

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
	log.Printf("ws request: exec pod:%s, container:%s, namespace:%s\n", pod, containerName, namespace)

	term, err := webshell.NewTerminalSession(w, r, nil)
	if term != nil {
		defer func() {
			log.Println("close session")
			term.Close()
		}()
	}
	if err != nil {
		log.Printf("get terminal session failed: %v\n", err)
		return
	}

	resource, err := createK8sResourceClient()
	if err != nil {
		msg := fmt.Sprintf("init k8s client error: %v\n", err)
		writeErrorRespToTerminal(term, msg)
		return
	}

	if containerName != "null" {
		if err := resource.CheckPodExec(r.Context(), namespace, pod, containerName); err != nil {
			log.Printf("check pod failed: pod:%s, container:%s, namespace:%s\n", pod, containerName, namespace)
			msg := fmt.Sprintf("validate pod error: %v\n", err)
			writeErrorRespToTerminal(term, msg)
			return
		}
	} else {
		pod, err := resource.GetPod(r.Context(), namespace, pod)
		if err != nil {
			msg := fmt.Sprintf("get pod [%s/%s] failed: %v\n", namespace, pod, err)
			writeErrorRespToTerminal(term, msg)
			return
		}
		containerName = pod.Spec.Containers[0].Name
	}

	config, err := k8sutils.GetK8sConfig()
	if err != nil {
		msg := fmt.Sprintf("get k8s config failed: %v\n", err)
		writeErrorRespToTerminal(term, msg)
		return
	}
	if err := webshell.ExecPod(resource.GetClient(), config, term, namespace, pod, containerName); err != nil {
		msg := fmt.Sprintf("exec pod error: %v\n", err)
		writeErrorRespToTerminal(term, msg)
	}
}

func writeErrorRespToTerminal(term *webshell.TerminalSession, msg string) {
	log.Println(msg)
	term.Write([]byte(msg))
	term.Done()

}

func writeInternalError(conn *websocket.Conn, msg string, err error) {
	err = fmt.Errorf("Internal server error.\nmessage: %s, error: %v", msg, err)
	conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
}

//
// http io
//

func writeOkJSONResp(w http.ResponseWriter, data []byte) {
	writeJSONRespWithStatus(w, http.StatusOK, data)
}

func writeJSONRespWithStatus(w http.ResponseWriter, retCode int, data []byte) {
	utils.AddCorsHeadersForOptions(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(retCode)

	if err := json.NewEncoder(w).Encode(respJSONData{Data: data}); err != nil {
		log.Println("Write json encoded response error:", err.Error())
	}
}
