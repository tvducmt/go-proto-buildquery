package buildquery

import (
	"fmt"
	"reflect"

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
		// if gogoproto.IsProto3(file.FileDescriptorProto) {
		b.generateProto3Message(file, msg)
		// }
	}
}

func (b *buildquery) getFieldQueryIfAny(field *descriptor.FieldDescriptorProto) *querier.FieldQuery {
	if field.Options != nil {
		v, err := proto.GetExtension(field.Options, querier.E_Field)
		if err == nil && v.(*querier.FieldQuery) != nil {
			return (v.(*querier.FieldQuery))
		}
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

		fieldQeurier := b.getFieldQueryIfAny(field)
		if fieldQeurier == nil {
			continue
		}
		fieldName := b.GetOneOfFieldName(message, field)
		variableName := "this." + fieldName
		b.P(generateStringQuerier(variableName, ccTypeName, fieldName, fieldQeurier))
		// }
	}
	b.P(`return query`)
	b.Out()
	b.P(`}`)
}
func isEnumAll(vv interface{}) bool {
	type enumInterface interface {
		EnumDescriptor() ([]byte, []int)
	}
	if _, ok := vv.(enumInterface); ok {
		return fmt.Sprintf("%d", vv) == "-1"
	}
	return false
}
func generateStringQuerier(variableName string, ccTypeName string, fieldName string, fv *querier.FieldQuery) string {
	// b.P(`fv.GetQuery() `, fv.GetQuery())
	switch fv.GetQuery() {
	case "=": //Term
		if reflect.TypeOf(variableName).Kind() == reflect.Slice {
			return `query = query.Filter(elastic.NewTermsQuery(params[0], DoubleSlice(vv)...))`
		} else if isEnumAll(variableName) {
			return `glog.Infoln(` + fieldName + `, Is enum all)`
		} else {
			//	comp := convertDateTimeSearch(vv, params[1])
			return `query = query.Filter(elastic.NewTermQuery(` + string(fieldName) + `,comp))`
		}
	case "mt":
		return `query = query.Must(elastic.NewMatchQuery(` + string(fieldName) + `,` + variableName + `))`
		// query = query.Must(elastic.NewMatchQuery(params[0], vv))
	case "match":
		return `query = query.Must(elastic.NewMatchQuery(params[0]+".search", vv).MinimumShouldMatch("3<90%"))`
	case ">=":
		return `glog.Infoln(params[0], vv)`
		// if !rangeDateSearch.addFrom(params[0], vv) {
		// 	query = query.Must(r.NewRangeQuery(params[0]).Gte(vv))
		// }

	default:
		return "nullll"
		// b.Out()
		//b.P(b.fmtPkg.Use(), `.Errorf("Unknow"`, fv.GetQuery(), `)`)

	}

}
