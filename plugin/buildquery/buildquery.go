package buildquery

import (
	"sync"

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
	elasticPkg generator.Single
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
	b.elasticPkg = b.NewImport("git.zapa.cloud/merchant-tools/helper/search/elastic")
	// if !gogoproto.ImportsGoGoProto(file.FileDescriptorProto) {
	// 	b.protoPkg = b.NewImport("github.com/golang/protobuf/proto")
	// }
	b.querierPkg = b.NewImport("github.com/tvducmt/go-proto-buildquery")

	b.P(`func convertDateTimeSearch(vv interface{}, op string) interface{} {`)
	b.P(`if date, ok := vv.(*`, b.protoPkg.Use(), `.Date); ok {`)
	b.P(`switch op {`)
	b.P(`case "<":`)
	b.P(`return `, b.protoPkg.Use(), `.DateUpperToTimeSearch(date).UnixNano() / int64(time.Millisecond)`)
	b.P(`default:`)
	b.P(`return `, b.protoPkg.Use(), `.DateToTimeSearch(date).UnixNano() / int64(time.Millisecond)`)
	b.P(`}`)
	b.P(`}`)
	b.P(`return vv`)
	b.P(`}`)

	b.P(`func isEnumAll(vv interface{}) bool {`)
	b.P(`type enumInterface interface {`)
	b.P(`EnumDescriptor() ([]byte, []int)`)
	b.P(`}`)
	b.P(`if _, ok := vv.(enumInterface); ok {`)
	b.P(`return fmt.Sprintf("%d", vv) == "-1"`)
	b.P(`}`)
	b.P(`return false`)
	b.P(`}`)

	b.P(`type rangeQuery struct {`)
	b.P(`mapQuery map[string]*`, b.elasticPkg.Use(), `.RangeQuery`)
	b.P(`}`)

	b.P(`func (r *rangeQuery) NewRangeQuery(name string) *`, b.elasticPkg.Use(), `.RangeQuery {`)
	b.P(`if q, ok := r.mapQuery[name]; ok {`)
	b.P(`return q`)
	b.P(`}`)
	b.P(`q := `, b.elasticPkg.Use(), `.NewRangeQuery(name)`)
	b.P(`r.mapQuery[name] = q`)
	b.P(`return q`)
	b.P(`}`)

	b.P(`type mapRangeDateSearch struct {`)
	b.P(`mapRangeDateSearch map[string]*rangeDateSearch`)
	b.P(`}`)

	b.P(`func dateToStringSearch(date *`, b.protoPkg.Use(), `.Date) string {`)
	b.P(`return fmt.Sprintf("%04d-%02d-%02d", date.GetYear(), date.GetMonth(), date.GetDay())`)
	b.P(`}`)

	b.P(`type rangeDateSearch struct {`)
	b.P(`from, to *`, b.protoPkg.Use(), `.Date`)
	b.P(`}`)

	b.P(`func (r *mapRangeDateSearch) addFrom(name string, vv interface{}) bool {`)
	b.P(`if from, ok := vv.(*`, b.protoPkg.Use(), `.Date); ok {`)
	b.P(`	if q, ok := r.mapRangeDateSearch[name]; ok {`)
	b.P(`		q.from = from`)
	b.P(`	} else {`)
	b.P(`		r.mapRangeDateSearch[name] = &rangeDateSearch{from: from}`)
	b.P(`	}`)
	b.P(`	return true`)
	b.P(`}`)
	b.P(`return false`)
	b.P(`}`)

	b.P(`func (r *mapRangeDateSearch) addTo(name string, vv interface{}) bool {`)
	b.P(`if to, ok := vv.(*`, b.protoPkg.Use(), `.Date); ok {`)
	b.P(`if q, ok := r.mapRangeDateSearch[name]; ok {`)
	b.P(`q.to = to`)
	b.P(`} else {`)
	b.P(`r.mapRangeDateSearch[name] = &rangeDateSearch{to: to}`)
	b.P(`	}`)
	b.P(`return true`)
	b.P(`}`)
	b.P(`return false`)
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
	b.P(`flag.Parse()`)
	b.P(`query := elastic.NewBoolQuery()`)
	b.In()
	rangeDateSearchDeclar := func() {
		b.P(`rangeDateSearch := &mapRangeDateSearch{mapRangeDateSearch: map[string]*rangeDateSearch{}}`)
		b.P(`bHasSearchPrefix, disableRangeFilter, searchPhone := false, false, false`)
	}

	once1 := &sync.Once{}
	once2 := &sync.Once{}
	for _, field := range message.Field {
		once1.Do(rangeDateSearchDeclar)
		fieldQeurier := b.getFieldQueryIfAny(field)
		if fieldQeurier == nil {
			continue
		}
		fieldName := b.GetOneOfFieldName(message, field)
		variableName := "this." + fieldName
		b.P(`fmt.Println(variableName, ` + variableName + `)`)
		// if variableName != "" {
		b.generateQuerier(once2, variableName, ccTypeName, fieldName, fieldQeurier)
		// }
		// }
	}
	b.P(`return query`)
	b.Out()
	b.P(`}`)
}

func (b *buildquery) generateQuerier(once *sync.Once, variableName string, ccTypeName string, fieldName string, fv *querier.FieldQuery) {
	rangeQueryDeclar := func() {
		b.P(`r := &rangeQuery{`)
		b.P(`mapQuery: map[string]*`, b.elasticPkg.Use(), `.RangeQuery{},`)
		b.P(`}`)

	}
	switch fv.GetQuery() {
	case "*%*":
		b.P(`bHasSearchPrefix = true`)
		b.P(`if !disableRangeFilter && len(fmt.Sprintf("%v",` + variableName + `)) >= 8 {`)
		b.P(`disableRangeFilter = true`)
		b.P(`}`)
		b.P(`if "` + fieldName + `" == "userInfo.phoneNumber" {`)
		b.P(`searchPhone = true`)
		b.P(`}`)
		b.P(`query = query.Must(elastic.NewMultiMatchQuery(`)
		b.P(variableName + `, "` + fieldName + `.search", "` + fieldName + `.search_reverse").MaxExpansions(1024).Slop(2).Type("phrase_prefix"))`)
	case "*%":
		b.P(`query = query.Must(elastic.NewMatchPhrasePrefixQuery(`)
		b.P(`fmt.Sprintf("%s.search", ` + fieldName + `),`)
		b.P(variableName + `,).MaxExpansions(1024).Slop(2))`)
	case "%*":
		b.P(`query = query.Must(elastic.NewMatchPhrasePrefixQuery(`)
		b.P(`fmt.Sprintf("%s.search_reverse", ` + fieldName + `,`)
		b.P(variableName + `,).MaxExpansions(1024).Slop(2))`)

	case "*.*": //Wildcard
		b.P(`s := fmt.Sprintf("%v", ` + variableName + `)`)
		b.P(`if !strings.Contains(s, "*") {`)
		b.P(`	s = "*" + s + "*"`)
		b.P(`}`)
		b.P(`query = query.Must(elastic.NewWildcardQuery(` + fieldName + `, s))`)
	case "*.": //Wildcard
		b.P(`query = query.Must(elastic.NewWildcardQuery(` + fieldName + `, fmt.Sprintf("*%v", ` + variableName + `)))`)
	case ".*": //Wildcard
		b.P(`query = query.Must(elastic.NewWildcardQuery(` + fieldName + `, fmt.Sprintf("%v*", ` + variableName + `)))`)
	case "=": //Term
		b.P(`if reflect.TypeOf(`, variableName, `).Kind() == reflect.Slice {`)
		b.P(`query = query.Filter(elastic.NewTermsQuery("` + fieldName + `",` + variableName + `))`)
		b.P(`} else if isEnumAll(`, variableName, `) {`)
		b.P(`glog.Infoln("` + fieldName + `", "Is enum all")`)
		b.P(`} else{`)
		b.P(`comp := convertDateTimeSearch(` + variableName + `,"=")`)
		b.P(`query = query.Filter(elastic.NewTermQuery("` + fieldName + `",comp))`)
		b.P(`}`)
	case "mt":
		b.P(`query = query.Must(elastic.NewMatchQuery("` + fieldName + `",` + variableName + `))`)
		// query = query.Must(elastic.NewMatchQuery(params[0], vv))
	case "match":
		b.P(`query = query.Must(elastic.NewMatchQuery("` + fieldName + `.search",` + variableName + `).MinimumShouldMatch("3<90%"))`)
	case ">=":
		b.P(`glog.Infoln("` + fieldName + `",` + variableName + `)`)
		once.Do(rangeQueryDeclar)
		b.P(`if !rangeDateSearch.addFrom("` + fieldName + `", ` + variableName + `) {`)
		b.P(`query = query.Must(r.NewRangeQuery("` + fieldName + `").Gte(` + variableName + `))`)
		b.P(`}`)
	case "<=":
		b.P(`glog.Infoln("` + fieldName + `",` + variableName + `)`)
		once.Do(rangeQueryDeclar)
		b.P(`if !rangeDateSearch.addTo("` + fieldName + `", ` + variableName + `) {`)
		b.P(`query = query.Must(r.NewRangeQuery("` + fieldName + `").Lte(` + variableName + `))`)
		b.P(`}`)
	case ">":
		b.P(`glog.Infoln("` + fieldName + `",` + variableName + `)`)
		once.Do(rangeQueryDeclar)
		b.P(`if !rangeDateSearch.addFrom("` + fieldName + `", ` + variableName + `) {`)
		b.P(`query = query.Must(r.NewRangeQuery("` + fieldName + `").Gt(` + variableName + `))`)
		b.P(`}`)
	case "<":
		b.P(`glog.Infoln("` + fieldName + `",` + variableName + `)`)
		once.Do(rangeQueryDeclar)
		b.P(`if !rangeDateSearch.addTo("` + fieldName + `", ` + variableName + `) {`)
		b.P(`	query = query.Must(r.NewRangeQuery("` + fieldName + `").Lt(` + variableName + `))`)
		b.P(`}`)
	case "!=":
		b.P(`glog.Infoln("` + fieldName + `",` + variableName + `)`)
		b.P(`if reflect.TypeOf(`, variableName, `).Kind() == reflect.Slice {`)
		b.P(`query = query.MustNot(elastic.NewTermsQuery("` + fieldName + `",` + variableName + `))`)
		b.P(`} else {`)
		b.P(`comp := convertDateTimeSearch(` + variableName + `,"!=")`)
		b.P(`query = query.MustNot(elastic.NewTermQuery("` + fieldName + `",comp))`)
		b.P(`}`)
	default:
		b.P(`glog.Warningln("Unknow ", params[1])`)
	}

	b.P(`if !disableRangeFilter || searchPhone {`)
	b.P(`for k, v := range rangeDateSearch.mapRangeDateSearch {`)
	b.P(`glog.Infoln(k, v)`)
	b.P(`f, t := v.from, v.to`)
	b.P(`if f != nil && t != nil {`)
	b.P(`if !searchPhone && bHasSearchPrefix && f.Day+7 > t.Day && f.Month == t.Month && f.Year == t.Year {`)
	b.P(`tm := time.Date(int(f.Year), time.Month(f.Month), int(f.Day), 0, 0, 0, 0, time.UTC).Add(-7 * 24 * time.Hour)`)
	b.P(`f.Year, f.Month, f.Day = int32(tm.Year()), int32(tm.Month()), int32(tm.Day())`)
	b.P(`}`)
	b.P(`query = query.Filter(elastic.NewRangeQuery(k).Gte(dateToStringSearch(f)).Lte(dateToStringSearch(t)).TimeZone("+07:00"))`)
	b.P(`} else {`)
	b.P(`glog.Errorln("Invalid ", k)`)
	b.P(`}`)

	b.P(`}`)
	b.P(`}`)

}
