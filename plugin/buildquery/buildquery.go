package buildquery

import (
	"fmt"

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
	b.protoPkg = b.NewImport("git.zapa.cloud/merchant-tools/helper/proto")
	// if !gogoproto.ImportsGoGoProto(file.FileDescriptorProto) {
	// 	b.protoPkg = b.NewImport("github.com/golang/protobuf/proto")
	// }
	b.querierPkg = b.NewImport("github.com/tvducmt/go-proto-buildquery")

	b.P(`func convertDateTimeSearch(vv interface{}, op string) interface{} {`)
	b.P(`if date, ok := vv.(*proto.Date); ok {`)
	b.P(`switch op {`)
	b.P(`case "<":`)
	b.P(`return proto.DateUpperToTimeSearch(date).UnixNano() / int64(time.Millisecond)`)
	b.P(`default:`)
	b.P(`return proto.DateToTimeSearch(date).UnixNano() / int64(time.Millisecond)`)
	b.P(`}`)
	b.P(`}`)
	b.P(`return vv`)
	b.P(`}`)

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
		b.P(`var `, fieldName, ` string`)
		variableName := "this." + fieldName
		b.generateStringQuerier(variableName, ccTypeName, fieldName, fieldQeurier)
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

func (b *buildquery) generateStringQuerier(variableName string, ccTypeName string, fieldName string, fv *querier.FieldQuery) {
	// b.P(`fv.GetQuery() `, fv.GetQuery())
	switch fv.GetQuery() {
	case "=": //Term
		b.P(`if reflect.TypeOf(`, variableName, `).Kind() == reflect.Slice {`)
		b.P(`query = query.Filter(elastic.NewTermsQuery(` + fieldName + `,` + variableName + `))`)
		b.P(`} else {`)
		b.P(`comp := convertDateTimeSearch(` + variableName + `,"=")`)
		b.P(`query = query.Filter(elastic.NewTermQuery(` + string(fieldName) + `,comp))`)
		b.P(`}`)
	case "mt":
		b.P(`query = query.Must(elastic.NewMatchQuery(` + string(fieldName) + `,` + variableName + `))`)
		// query = query.Must(elastic.NewMatchQuery(params[0], vv))
	case "match":
		b.P(`query = query.Must(elastic.NewMatchQuery(params[0]+".search", vv).MinimumShouldMatch("3<90%"))`)
	case ">=":
		b.P(`glog.Infoln(params[0], vv)`)
		// if !rangeDateSearch.addFrom(params[0], vv) {
		// 	query = query.Must(r.NewRangeQuery(params[0]).Gte(vv))
		// }

	default:
		b.P("nullll")
		// b.Out()
		//b.P(b.fmtPkg.Use(), `.Errorf("Unknow"`, fv.GetQuery(), `)`)

	}

}
