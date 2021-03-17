package internal

import (
	"encoding/json"
	"fmt"

	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
)

// IsMethodValidate returns whether method is in method list of grpc service.
func IsMethodValidate(descSource grpcurl.DescriptorSource, methodName string) (bool, error) {
	allServices, err := descSource.ListServices()
	if err != nil {
		return false, err
	}

	for _, svc := range allServices {
		if svc == "grpc.reflection.v1alpha.ServerReflection" {
			continue
		}
		d, err := descSource.FindSymbol(svc)
		if err != nil {
			return false, err
		}
		sd, ok := d.(*desc.ServiceDescriptor)
		if !ok {
			return false, fmt.Errorf("%s cannot convert to ServiceDescriptor", sd.GetFullyQualifiedName())
		}

		for _, md := range sd.GetMethods() {
			if methodName == md.GetFullyQualifiedName() {
				return true, nil
			}
		}
	}
	return false, nil
}

// PrintGrpcServiceInfo prints grpc services info.
func PrintGrpcServiceInfo(descSource grpcurl.DescriptorSource) error {
	allServices, err := descSource.ListServices()
	if err != nil {
		return err
	}

	for _, svc := range allServices {
		if svc == "grpc.reflection.v1alpha.ServerReflection" {
			continue
		}
		fmt.Println("* service", svc)
		d, err := descSource.FindSymbol(svc)
		if err != nil {
			return err
		}
		sd, ok := d.(*desc.ServiceDescriptor)
		if !ok {
			return fmt.Errorf("%s cannot convert to ServiceDescriptor", sd.GetFullyQualifiedName())
		}

		for _, md := range sd.GetMethods() {
			fmt.Println("\trpc method:", md.GetFullyQualifiedName())
			msgDesc := md.GetInputType()
			fmt.Println("\t\tmessage:", msgDesc.GetFullyQualifiedName())
			for _, fd := range msgDesc.GetFields() {
				fmt.Println("\t\t\tfield:", fd.GetFullyQualifiedName(), fd.GetType())
			}

			meta, err := getMetadataForMethod(md)
			if err != nil {
				return err
			}
			// b, err := json.Marshal(meta)
			b, err := json.MarshalIndent(meta, "", "\t")
			if err != nil {
				return err
			}
			fmt.Println("\nmethod metadata:", string(b))
			fmt.Println()
		}
	}
	return nil
}

/*
Method Metadata

Refer: grpcui/handler.go
*/

const (
	typeString   fieldType = "string"
	typeBytes    fieldType = "bytes"
	typeInt32    fieldType = "int32"
	typeInt64    fieldType = "int64"
	typeSint32   fieldType = "sint32"
	typeSint64   fieldType = "sint64"
	typeUint32   fieldType = "uint32"
	typeUint64   fieldType = "uint64"
	typeFixed32  fieldType = "fixed32"
	typeFixed64  fieldType = "fixed64"
	typeSfixed32 fieldType = "sfixed32"
	typeSfixed64 fieldType = "sfixed64"
	typeFloat    fieldType = "float"
	typeDouble   fieldType = "double"
	typeBool     fieldType = "bool"
	typeOneOf    fieldType = "oneof"
)

var typeMap = map[descriptor.FieldDescriptorProto_Type]fieldType{
	descriptor.FieldDescriptorProto_TYPE_STRING:   typeString,
	descriptor.FieldDescriptorProto_TYPE_BYTES:    typeBytes,
	descriptor.FieldDescriptorProto_TYPE_INT32:    typeInt32,
	descriptor.FieldDescriptorProto_TYPE_INT64:    typeInt64,
	descriptor.FieldDescriptorProto_TYPE_SINT32:   typeSint32,
	descriptor.FieldDescriptorProto_TYPE_SINT64:   typeSint64,
	descriptor.FieldDescriptorProto_TYPE_UINT32:   typeUint32,
	descriptor.FieldDescriptorProto_TYPE_UINT64:   typeUint64,
	descriptor.FieldDescriptorProto_TYPE_FIXED32:  typeFixed32,
	descriptor.FieldDescriptorProto_TYPE_FIXED64:  typeFixed64,
	descriptor.FieldDescriptorProto_TYPE_SFIXED32: typeSfixed32,
	descriptor.FieldDescriptorProto_TYPE_SFIXED64: typeSfixed64,
	descriptor.FieldDescriptorProto_TYPE_FLOAT:    typeFloat,
	descriptor.FieldDescriptorProto_TYPE_DOUBLE:   typeDouble,
	descriptor.FieldDescriptorProto_TYPE_BOOL:     typeBool,
}

type schema struct {
	RequestType   string                  `json:"requestType"`
	RequestStream bool                    `json:"requestStream"`
	MessageTypes  map[string][]fieldDef   `json:"messageTypes"`
	EnumTypes     map[string][]enumValDef `json:"enumTypes"`
}

type fieldDef struct {
	Name        string      `json:"name"`
	ProtoName   string      `json:"protoName"`
	Type        fieldType   `json:"type"`
	OneOfFields []fieldDef  `json:"oneOfFields"`
	IsMessage   bool        `json:"isMessage"`
	IsEnum      bool        `json:"isEnum"`
	IsArray     bool        `json:"isArray"`
	IsMap       bool        `json:"isMap"`
	IsRequired  bool        `json:"isRequired"`
	DefaultVal  interface{} `json:"defaultVal"`
}

type enumValDef struct {
	Num  int32  `json:"num"`
	Name string `json:"name"`
}

type fieldType string

func getMetadataForMethod(md *desc.MethodDescriptor) (*schema, error) {
	msg := md.GetInputType()
	result := &schema{
		RequestType:   msg.GetFullyQualifiedName(),
		RequestStream: md.IsClientStreaming(),
		MessageTypes:  map[string][]fieldDef{},
		EnumTypes:     map[string][]enumValDef{},
	}

	result.visitMessage(msg)
	return result, nil
}

func (s *schema) visitMessage(md *desc.MessageDescriptor) {
	if _, ok := s.MessageTypes[md.GetFullyQualifiedName()]; ok {
		// already visited
		return
	}

	fields := make([]fieldDef, 0, len(md.GetFields()))
	s.MessageTypes[md.GetFullyQualifiedName()] = fields

	oneOfsSeen := map[*desc.OneOfDescriptor]struct{}{}
	for _, fd := range md.GetFields() {
		ood := fd.GetOneOf()
		if ood != nil {
			if _, ok := oneOfsSeen[ood]; ok {
				// already processed this one
				continue
			}
			oneOfsSeen[ood] = struct{}{}
			fields = append(fields, s.processOneOf(ood))
		} else {
			fields = append(fields, s.processField(fd))
		}
	}

	s.MessageTypes[md.GetFullyQualifiedName()] = fields
}

func (s *schema) processOneOf(ood *desc.OneOfDescriptor) fieldDef {
	choices := make([]fieldDef, len(ood.GetChoices()))
	for i, fd := range ood.GetChoices() {
		choices[i] = s.processField(fd)
	}
	return fieldDef{
		Name:        ood.GetName(),
		Type:        typeOneOf,
		OneOfFields: choices,
	}
}

func (s *schema) processField(fd *desc.FieldDescriptor) fieldDef {
	def := fieldDef{
		Name:       fd.GetJSONName(),
		ProtoName:  fd.GetName(),
		IsEnum:     fd.GetEnumType() != nil,
		IsMessage:  fd.GetMessageType() != nil,
		IsArray:    fd.IsRepeated() && !fd.IsMap(),
		IsMap:      fd.IsMap(),
		IsRequired: fd.IsRequired(),
		DefaultVal: fd.GetDefaultValue(),
	}

	if def.IsMap {
		// fd.GetDefaultValue returns empty map[interface{}]interface{}
		// as the default for map fields, but "encoding/json" refuses
		// to encode a map with interface{} keys (even if it's empty).
		// So we fix up the key type here.
		def.DefaultVal = map[string]interface{}{}
	}

	// 64-bit int values are represented as strings in JSON
	if i, ok := def.DefaultVal.(int64); ok {
		def.DefaultVal = fmt.Sprintf("%d", i)
	} else if u, ok := def.DefaultVal.(uint64); ok {
		def.DefaultVal = fmt.Sprintf("%d", u)
	} else if b, ok := def.DefaultVal.([]byte); ok && b == nil {
		// bytes fields may have []byte(nil) as default value, but
		// that gets rendered as JSON null, not empty array
		def.DefaultVal = []byte{}
	}

	switch fd.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		def.Type = fieldType(fd.GetEnumType().GetFullyQualifiedName())
		s.visitEnum(fd.GetEnumType())
		// DefaultVal will be int32 for enums, but we want to instead
		// send enum name as string
		if val, ok := def.DefaultVal.(int32); ok {
			valDesc := fd.GetEnumType().FindValueByNumber(val)
			if valDesc != nil {
				def.DefaultVal = valDesc.GetName()
			}
		}

	case descriptor.FieldDescriptorProto_TYPE_GROUP, descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		def.Type = fieldType(fd.GetMessageType().GetFullyQualifiedName())
		s.visitMessage(fd.GetMessageType())

	default:
		def.Type = typeMap[fd.GetType()]
	}

	return def
}

func (s *schema) visitEnum(ed *desc.EnumDescriptor) {
	if _, ok := s.EnumTypes[ed.GetFullyQualifiedName()]; ok {
		// already visited
		return
	}

	enumVals := make([]enumValDef, len(ed.GetValues()))
	for i, evd := range ed.GetValues() {
		enumVals[i] = enumValDef{
			Num:  evd.GetNumber(),
			Name: evd.GetName(),
		}
	}

	s.EnumTypes[ed.GetFullyQualifiedName()] = enumVals
}
