package service

import (
	"go1_1711_demo/ioc.demo/autowire_rpc/server/pkg/dto"
)

// +ioc:autowire=true
// +ioc:autowire:type=rpc

type ServiceStruct struct {
}

func (s *ServiceStruct) GetUser(name string, age int) (*dto.User, error) {
	return &dto.User{
		Id:   1,
		Name: name,
		Age:  age,
	}, nil
}
