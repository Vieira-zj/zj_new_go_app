package main

import (
	"context"
	"crypto/tls"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"demo.hello/k8s/webhook/pkg"
	"github.com/golang/glog"
)

type whSvrParameters struct {
	Port           int    // webhook server port
	CertFile       string // path to the x509 certificate for https
	KeyFile        string // path to the x509 private key matching `CertFile`
	sidecarCfgFile string // path to sidecar injector configuration file
}

func main() {
	var parameters whSvrParameters

	flag.IntVar(&parameters.Port, "port", 443, "Webhook server port.")
	flag.StringVar(&parameters.CertFile, "tlsCertFile", "/etc/webhook/certs/cert.pem", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&parameters.KeyFile, "tlsKeyFile", "/etc/webhook/certs/key.pem", "File containing the x509 private key to --tlsCertFile.")
	flag.StringVar(&parameters.sidecarCfgFile, "sidecarCfgFile", "/etc/webhook/config/sidecarconfig.yaml", "File containing the mutation configuration.")

	help := flag.Bool("help", false, "Help.")
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	pair, err := tls.LoadX509KeyPair(parameters.CertFile, parameters.KeyFile)
	if err != nil {
		glog.Errorf("Failed to load key pair: %v", err)
		return
	}

	// define http server and server handler
	whsvr := pkg.NewWebhookServer(parameters.sidecarCfgFile, parameters.Port, pair)
	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", whsvr.Serve)
	mux.HandleFunc("/validate", whsvr.Serve)
	mux.HandleFunc("/inject", whsvr.Serve)
	whsvr.Server.Handler = mux

	// start webhook server in new routine
	go func() {
		if err := whsvr.Server.ListenAndServeTLS("", ""); err != nil {
			glog.Errorf("Failed to listen and serve webhook server: %v", err)
		}
	}()
	glog.Info("Server started")

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	glog.Infof("Got OS shutdown signal, shutting down webhook server gracefully...")
	whsvr.Server.Shutdown(context.Background())
}
