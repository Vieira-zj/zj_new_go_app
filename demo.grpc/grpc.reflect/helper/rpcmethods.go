package helper

import (
	"context"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"

	rpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

// AllMethodsViaReflection returns a slice that contains the method descriptors
// for each method exposed by the server.
func AllMethodsViaReflection(ctx context.Context, cc grpc.ClientConnInterface) (map[string][]*desc.MethodDescriptor, error) {
	stub := rpb.NewServerReflectionClient(cc)
	client := grpcreflect.NewClient(ctx, stub)
	svcNames, err := client.ListServices()
	if err != nil {
		return nil, err
	}

	var descs []*desc.ServiceDescriptor
	for _, svcName := range svcNames {
		sd, err := client.ResolveService(svcName)
		if err != nil {
			return nil, err
		}
		// skip reflection service
		if sd.GetFullyQualifiedName() == "grpc.reflection.v1alpha.ServerReflection" {
			continue
		}
		descs = append(descs, sd)
	}
	return allMethodsForServices(descs), nil
}

// AllMethodsForServer returns a slice that contains the method descriptors for
// each method exposed by the given gRPC server.
func AllMethodsForServer(svr *grpc.Server) (map[string][]*desc.MethodDescriptor, error) {
	svcdescs, err := grpcreflect.LoadServiceDescriptors(svr)
	if err != nil {
		return nil, err
	}

	var descs []*desc.ServiceDescriptor
	for _, sd := range svcdescs {
		descs = append(descs, sd)
	}
	return allMethodsForServices(descs), nil
}

// allMethodsForServices returns a slice that contains the method descriptors
// for each method in the given services.
func allMethodsForServices(svcdescs []*desc.ServiceDescriptor) map[string][]*desc.MethodDescriptor {
	ret := make(map[string][]*desc.MethodDescriptor)
	for _, sd := range svcdescs {
		name := sd.GetFullyQualifiedName()
		if _, ok := ret[name]; ok {
			continue
		}
		ret[name] = sd.GetMethods()
	}
	return ret
}
