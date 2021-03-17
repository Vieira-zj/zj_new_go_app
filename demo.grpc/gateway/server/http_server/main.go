package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/labstack/echo"
	"google.golang.org/grpc"

	gw "demo.grpc/gateway/proto/demo/hello"

	"demo.grpc/gateway/handlers"
)

/*
Config
*/

var (
	// gRPC server endpoint
	grpcServerEndpoint = flag.String("grpc-endpoint", "localhost:9090", "gRPC server endpoint")
	gatewayPort        = flag.String("gateway-port", ":8081", "gRPC gateway server port")
	mysqlPort          = flag.String("mysql-port", ":9091", "mysql proxy server port")
)

/*
Request and Response models
*/

type requestJSON struct {
	Query string `json:"query"`
}

type responseMeta struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type responseData struct {
	Results []string `json:"results"`
}

type responseJSON struct {
	Meta responseMeta `json:"meta"`
	Data responseData `json:"data"`
}

/*
Echo Handlers
*/

func queryHandler(c echo.Context) error {
	bodyBytes, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	request := &requestJSON{}
	if err := json.Unmarshal(bodyBytes, &request); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	handler := handlers.NewMysqlHandler()
	res, err := handler.ExecSelect(request.Query)

	var respMeta responseMeta
	var respData responseData
	if err != nil {
		respMeta.Code = -1
		respMeta.Message = err.Error()
	} else {
		respMeta.Code = 0
		respMeta.Message = "success"
		respData.Results = res
	}
	resp := &responseJSON{
		Meta: respMeta,
		Data: respData,
	}
	return c.JSON(http.StatusOK, resp)
}

/*
Process
*/

func customMatcher(key string) (string, bool) {
	switch key {
	case "X-User-Id":
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(customMatcher))
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := gw.RegisterService1HandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return err
	}
	err = gw.RegisterService2HandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return err
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	log.Printf("start grpc gateway server (proxy to %s) at: %s\n", *grpcServerEndpoint, *gatewayPort)
	return http.ListenAndServe(*gatewayPort, mux)
}

func runMysql() {
	e := echo.New()
	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.POST("/db/query", queryHandler)

	log.Println("start mysql proxy server listen at:", *mysqlPort)
	e.Logger.Fatal(e.Start((*mysqlPort)))
}

func main() {
	flag.Parse()
	defer glog.Flush()

	go runMysql()

	if err := run(); err != nil {
		glog.Fatal(err)
	}
}
