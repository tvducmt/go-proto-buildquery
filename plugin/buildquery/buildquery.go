package buildquery

import (
	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	querier "github.com/tvducmt/protoc-gen-buildquery/protobuf"
)

type buildquery struct {
	*generator.Generator
	generator.PluginImports
	querierPkg generator.Single
	fmtPkg     generator.Single
	protoPkg   generator.Single
	// query *elastic.BoolQuery
}

// NewBuildquery ...
func NewBuildquery() generator.Plugin {
	return &buildquery{
		// query: query,
	}
}

func (b *buildquery) Name() string {
	return "buildquery"
}

func (b *buildquery) Init(g *generator.Generator) {
	b.Generator = g
}

func (b *buildquery) Generate(file *generator.FileDescriptor) {
	// proto3 := gogoproto.IsProto3(file.FileDescriptorProto)
	b.PluginImports = generator.NewPluginImports(b.Generator)

	b.fmtPkg = b.NewImport("fmt")
	// stringsPkg := b.NewImport("strings")
	b.protoPkg = b.NewImport("github.com/gogo/protobuf/proto")
	if !gogoproto.ImportsGoGoProto(file.FileDescriptorProto) {
		b.protoPkg = b.NewImport("github.com/golang/protobuf/proto")
	}
	b.querierPkg = b.NewImport("github.com/tvducmt/go-proto-buildquery")

	for _, msg := range file.Messages() {
		if msg.DescriptorProto.GetOptions().GetMapEntry() {
			continue
		}
		// b.generateRegexVars(file, msg)
		if gogoproto.IsProto3(file.FileDescriptorProto) {
			b.generateProto3Message(file, msg)
		}
	}
}

func (b *buildquery) getFieldQueryIfAny(field *descriptor.FieldDescriptorProto) *querier.FieldQuery {
	if field.Options != nil {
		v, err := proto.GetExtension(field, querier.E_Field)
		if err == nil && v.(*querier.FieldQuery) != nil {
			b.P(`querier.FieldQuery`)
			return (v.(*querier.FieldQuery))
		}
		b.P(`field.OptionsError`, err.Error())

	}
	return nil
}
func (b *buildquery) generateProto3Message(file *generator.FileDescriptor, message *generator.Descriptor) {
	ccTypeName := generator.CamelCaseSlice(message.TypeName())
	b.P(`func (this *`, ccTypeName, `) BuildQuery() *elastic.BoolQuery {`)
	b.In()
	b.P(`query := elastic.NewBoolQuery()`)
	b.In()
	for _, field := range message.Field {
		b.P(`field`, field.GetName())
		//fieldQeurier := b.getFieldQueryIfAny(field)
		//b.P(`fieldQeurier`, fieldQeurier.GetQuery())
		// if fieldQeurier == nil {
		// 	continue
		// }
		fieldName := b.GetOneOfFieldName(message, field)
		// variableName := "this." + fieldName
		b.P(`fieldName`, fieldName)
		b.P(`field.IsString()`, field.IsString())
		//if field.IsString() {
		//b.generateStringQuerier(variableName, ccTypeName, fieldName, fieldQeurier)
		// }
	}
	b.P(`return query`)
	b.Out()
	b.P(`}`)
}
func (b *buildquery) generateStringQuerier(variableName string, ccTypeName string, fieldName string, fv *querier.FieldQuery) {
	b.P(`fv.GetQuery() `, fv.GetQuery())
	switch fv.GetQuery() {
	case "mt":
		b.Out()
		b.P(`query = query.Must(elastic.NewMatchQuery(`, fieldName, `,`, ccTypeName, `.fieldName))`)

	default:
		b.Out()
		b.P(b.fmtPkg.Use(), `.Errorf("Unknow"`, fv.GetQuery(), `)`)

	}

}
