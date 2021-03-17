package pkg

import (
	"fmt"

	"github.com/emicklei/proto"
)

// TestVisitor visit for proto elements.
type TestVisitor struct {
}

// VisitMessage visit message elements.
func (t TestVisitor) VisitMessage(m *proto.Message) {}

// VisitService visit service elements.
func (t TestVisitor) VisitService(v *proto.Service) {}

// VisitSyntax visit syntax elements.
func (t TestVisitor) VisitSyntax(s *proto.Syntax) {}

// VisitPackage visit pakcage elements.
func (t TestVisitor) VisitPackage(p *proto.Package) {}

// VisitOption visit option elements.
func (t TestVisitor) VisitOption(o *proto.Option) {}

// VisitImport visit import elements.
func (t TestVisitor) VisitImport(i *proto.Import) {}

// VisitNormalField visit normal field elements.
func (t TestVisitor) VisitNormalField(i *proto.NormalField) {
	fmt.Printf("\tnormal field: name=%s, optional=%v, sequence=%d\n", i.Name, i.Optional, i.Sequence)
}

// VisitEnumField visit enum field elements.
func (t TestVisitor) VisitEnumField(i *proto.EnumField) {}

// VisitEnum visit enum elements.
func (t TestVisitor) VisitEnum(e *proto.Enum) {}

// VisitComment visit comment elements.
func (t TestVisitor) VisitComment(e *proto.Comment) {}

// VisitOneof visit oneof elements.
func (t TestVisitor) VisitOneof(o *proto.Oneof) {}

// VisitOneofField visit oneof field elements.
func (t TestVisitor) VisitOneofField(o *proto.OneOfField) {}

// VisitReserved visit reserved elements.
func (t TestVisitor) VisitReserved(r *proto.Reserved) {}

// VisitRPC visit rpc elements.
func (t TestVisitor) VisitRPC(r *proto.RPC) {
	fmt.Println("\trpc:", r.Name)
}

// VisitMapField visit map field elements.
func (t TestVisitor) VisitMapField(f *proto.MapField) {}

// VisitGroup visit group elements.
func (t TestVisitor) VisitGroup(g *proto.Group) {}

// VisitExtensions visit extensions elements.
func (t TestVisitor) VisitExtensions(e *proto.Extensions) {}
