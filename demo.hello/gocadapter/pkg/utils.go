package pkg

import (
	"context"
	"fmt"
	"log"
	"time"
)

// IsAttachServerOK checks whether attach server is ok, and retry 3 times default.
func IsAttachServerOK(ctx context.Context, addr string) bool {
	for i := 1; ; i++ {
		if _, err := GetServiceCoverage(ctx, addr); err != nil {
			if i >= 4 {
				return false
			}
			log.Printf("IsAttachServerOK get service [%s] coverage failed: %s, retry %d", addr, err, i)
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			return true
		}
	}
}

// RemoveUnhealthServicesFromGocSvrList removes unhealth service from goc register services list.
func RemoveUnhealthServicesFromGocSvrList(ctx context.Context, host string) error {
	goc := NewGocAPI(host)
	services, err := goc.ListRegisterServices(ctx)
	if err != nil {
		return fmt.Errorf("RemoveUnhealthServicesFromGocSvrList get goc register service list failed: %w", err)
	}

	for _, addrs := range services {
		for _, addr := range addrs {
			if !IsAttachServerOK(ctx, addr) {
				if _, err := goc.DeleteRegisterServiceByAddr(ctx, addr); err != nil {
					return fmt.Errorf("RemoveUnhealthServicesFromGocSvrList remove goc register service [%s] failed: %w", addr, err)
				}
			}
		}
	}
	return nil
}
