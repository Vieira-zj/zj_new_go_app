package pkg

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"demo.hello/utils"
)

const (
	testGocLocalHost = "http://localhost:7777"
)

func TestNewGocAPIOnce(t *testing.T) {
	AppConfig.GocCenterIngHost = testGocLocalHost
	for i := 0; i < 3; i++ {
		NewGocAPI()
	}
}

func TestListRegisterServices(t *testing.T) {
	AppConfig.GocCenterIngHost = testGocLocalHost
	goc := NewGocAPI()
	svc, err := goc.ListRegisterServices(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("services: %+v\n", svc)
}

func TestRegisterService(t *testing.T) {
	AppConfig.GocCenterIngHost = testGocLocalHost
	goc := NewGocAPI()
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
	AppConfig.GocCenterIngHost = testGocLocalHost
	goc := NewGocAPI()
	service := "test_th_apa_goc_echoserver_v1"
	resp, err := goc.DeleteRegisterServiceByName(context.Background(), service)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp)
}

func TestDeleteRegisterServiceByAddr(t *testing.T) {
	AppConfig.GocCenterIngHost = testGocLocalHost
	goc := NewGocAPI()
	addr := "http://127.0.0.1:49971"
	resp, err := goc.DeleteRegisterServiceByAddr(context.Background(), addr)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp)
}

func TestGetServiceProfileByAddr(t *testing.T) {
	// curl http://localhost:8081/
	AppConfig.GocCenterIngHost = testGocLocalHost
	goc := NewGocAPI()
	addr := "http://127.0.0.1:51007"
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

func TestGetServiceProfileByAddrNotFound(t *testing.T) {
	AppConfig.GocCenterIngHost = testGocLocalHost
	goc := NewGocAPI()
	addr := "http://127.0.0.1:51027"
	b, err := goc.GetServiceProfileByAddr(context.Background(), addr)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(b))
}

func TestGetServiceProfileByName(t *testing.T) {
	// if service has multiple addresses (instance), goc will merge profile and return.
	AppConfig.GocCenterIngHost = testGocLocalHost
	srvName := "staging_th_apa_goc_echoserver_master_b63d82705a"
	goc := NewGocAPI()
	b, err := goc.GetServiceProfileByName(context.Background(), srvName)
	if err != nil {
		t.Fatal(err)
	}

	path := "/tmp/test/apa_goc_echoserver/repo/srv_profile.cov"
	if err = utils.CreateFile(path, b); err != nil {
		t.Fatal(err)
	}
}

func TestClearServiceProfileByAddr(t *testing.T) {
	AppConfig.GocCenterIngHost = testGocLocalHost
	goc := NewGocAPI()
	addr := "http://127.0.0.1:51025"
	resp, err := goc.ClearServiceProfileByAddr(context.Background(), addr)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp)
}

func TestClearProfileServiceByName(t *testing.T) {
	AppConfig.GocCenterIngHost = testGocLocalHost
	srvName := "staging_th_apa_goc_echoserver_master_b63d82705a"
	goc := NewGocAPI()
	resp, err := goc.ClearProfileServiceByName(context.Background(), srvName)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp)
}

func TestGetServiceCoverage(t *testing.T) {
	AppConfig.GocCenterIngHost = testGocLocalHost
	goc := NewGocAPI()
	services, err := goc.ListRegisterServices(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	for _, addrs := range services {
		if len(addrs) > 0 {
			addr := addrs[0]
			cov, err := APIGetServiceCoverage(context.Background(), addr)
			if err != nil {
				t.Fatal(err)
			}
			total, err := formatCoverPercentage(cov)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Printf("address:[%s], coverage:%s\n", addr, total)
		}
	}
}
