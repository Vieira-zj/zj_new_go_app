package pkg

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
)

const (
	testGocLocalHost = "http://localhost:7777"
)

func TestNewGocAPIOnce(t *testing.T) {
	for i := 0; i < 3; i++ {
		NewGocAPI(testGocLocalHost)
	}
}

func TestListRegisterServices(t *testing.T) {
	goc := NewGocAPI(testGocLocalHost)
	svc, err := goc.ListRegisterServices(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("services: %+v\n", svc)
}

func TestRegisterService(t *testing.T) {
	goc := NewGocAPI(testGocLocalHost)
	for i := 1; i < 3; i++ {
		service := "test_th_apa_goc_echoserver_v1"
		addr := fmt.Sprintf("http://127.0.0.1:4997%d", i)
		resp, err := goc.RegisterService(context.Background(), service, addr)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(resp)
	}
}

func TestDeleteRegisterServiceByName(t *testing.T) {
	goc := NewGocAPI(testGocLocalHost)
	service := "test_th_apa_goc_echoserver_v1"
	resp, err := goc.DeleteRegisterServiceByName(context.Background(), service)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp)
}

func TestDeleteRegisterServiceByAddr(t *testing.T) {
	goc := NewGocAPI(testGocLocalHost)
	addr := "http://127.0.0.1:49971"
	resp, err := goc.DeleteRegisterServiceByAddr(context.Background(), addr)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp)
}

func TestGetServiceProfileByAddr(t *testing.T) {
	// curl http://localhost:8081/
	goc := NewGocAPI(testGocLocalHost)
	addr := "http://127.0.0.1:51025"
	profile, err := goc.GetServiceProfileByAddr(context.Background(), addr)
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create("/tmp/test/goc.cov")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	buf := bytes.NewBuffer(profile)
	if _, err := f.Write(buf.Bytes()); err != nil {
		t.Fatal()
	}
	fmt.Println("get profile done")
}

func TestClearServiceProfileByAddr(t *testing.T) {
	goc := NewGocAPI(testGocLocalHost)
	addr := "http://127.0.0.1:51025"
	resp, err := goc.ClearServiceProfileByAddr(context.Background(), addr)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp)
}

func TestGetServiceCoverage(t *testing.T) {
	goc := NewGocAPI(testGocLocalHost)
	services, err := goc.ListRegisterServices(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	for _, hosts := range services {
		if len(hosts) > 0 {
			host := hosts[0]
			cov, err := APIGetServiceCoverage(context.Background(), host)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Printf("host [%s], coverage: %s\n", host, cov)
		}
	}
}