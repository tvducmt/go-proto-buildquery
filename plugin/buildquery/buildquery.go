package buildquery

import (
	"strings"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/golang/glog"
	querier "github.com/tvducmt/protoc-gen-buildquery/protobuf"
)

type buildquery struct {
	*generator.Generator
	generator.PluginImports
	querierPkg generator.Single
	glogPkg    generator.Single
	protoPkg   generator.Single
	elasticPkg generator.Single
	reflectPkg generator.Single
	timePkg    generator.Single
	flagPkg    generator.Single
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

	b.glogPkg = b.NewImport("github.com/golang/glog")
	// stringsPkg := b.NewImport("strings")
	b.protoPkg = b.NewImport("git.zapa.cloud/merchant-tools/helper/proto")
	b.elasticPkg = b.NewImport("git.zapa.cloud/merchant-tools/helper/search/elastic")
	b.reflectPkg = b.NewImport("reflect")
	b.timePkg = b.NewImport("time")
	b.flagPkg = b.NewImport("flag")
	// if !gogoproto.ImportsGoGoProto(file.FileDescriptorProto) {
	// 	b.protoPkg = b.NewImport("github.com/golang/protobuf/proto")
	// }
	b.querierPkg = b.NewImport("github.com/tvducmt/go-proto-buildquery")

	b.P(`func convertDateTimeSearch(vv interface{}, op string) interface{} {`)
	b.P(`if date, ok := vv.(*`, b.protoPkg.Use(), `.Date); ok {`)
	b.P(`switch op {`)
	b.P(`case "<":`)
	b.P(`return `, b.protoPkg.Use(), `.DateUpperToTimeSearch(date).UnixNano() / int64(`, b.timePkg.Use(), `.Millisecond)`)
	b.P(`default:`)
	b.P(`return `, b.protoPkg.Use(), `.DateToTimeSearch(date).UnixNano() / int64(`, b.timePkg.Use(), `.Millisecond)`)
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

	b.P(`func checkNull(field interface{}) bool {`)
	b.P(`zero := reflect.Zero(reflect.TypeOf(field)).Interface()	`)
	b.P(`if reflect.DeepEqual(field, zero) {`)
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
	b.P(`func (this *`, ccTypeName, `) BuildQuery(query *`, b.elasticPkg.Use(), `.BoolQuery) *`, b.elasticPkg.Use(), `.BoolQuery {`)
	b.In()
	b.P(b.flagPkg.Use(), `.Parse()`)
	b.P(`if query == nil {`)
	b.P(`query = `, b.elasticPkg.Use(), `.NewBoolQuery()`)
	b.P(`}`)
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

		// b.P(`fmt.Println("variableName", ` + variableName + `)`)
		b.generateQuerier(once2, variableName, ccTypeName, fieldQeurier)
	}
	b.P(`if !disableRangeFilter || searchPhone {`)
	b.P(`for k, v := range rangeDateSearch.mapRangeDateSearch {`)
	b.P(b.glogPkg.Use(), `.Infoln(k, v)`)
	b.P(`f, t := v.from, v.to`)
	b.P(`if f != nil && t != nil {`)
	b.P(`if !searchPhone && bHasSearchPrefix && f.Day+7 > t.Day && f.Month == t.Month && f.Year == t.Year {`)
	b.P(`tm := `, b.timePkg.Use(), `.Date(int(f.Year), `, b.timePkg.Use(), `.Month(f.Month), int(f.Day), 0, 0, 0, 0, `, b.timePkg.Use(), `.UTC).Add(-7 * 24 * `, b.timePkg.Use(), `.Hour)`)
	b.P(`f.Year, f.Month, f.Day = int32(tm.Year()), int32(tm.Month()), int32(tm.Day())`)
	b.P(`}`)
	b.P(`query = query.Filter(`, b.elasticPkg.Use(), `.NewRangeQuery(k).Gte(dateToStringSearch(f)).Lte(dateToStringSearch(t)).TimeZone("+07:00"))`)
	b.P(`} else {`)
	b.P(b.glogPkg.Use(), `.Errorln("Invalid ", k)`)
	b.P(`}`)
	b.P(`}`)
	b.P(`}`)

	b.P(`return query`)
	b.Out()
	b.P(`}`)
}

func (b *buildquery) generateQuerier(once *sync.Once, variableName string, ccTypeName string, fv *querier.FieldQuery) {

	rangeQueryDeclar := func() {
		b.P(`r := &rangeQuery{`)
		b.P(`mapQuery: map[string]*`, b.elasticPkg.Use(), `.RangeQuery{},`)
		b.P(`}`)

	}
	tag := fv.GetQuery()
	params := strings.Split(tag, ",")
	if len(params) != 2 {
		glog.Warningln(tag, len(params))
		return
	}

	switch params[1] {
	case "*%*":
		b.P(`if !checkNull( ` + variableName + `){`)

		b.P(`fields := strings.Split("`, params[0], `", ";")`)
		b.P(`if len(fields) < 2 {`)
		b.P(`bHasSearchPrefix = true`)
		b.P(`if !disableRangeFilter && len(fmt.Sprintf("%v",` + variableName + `)) >= 8 {`)
		b.P(`disableRangeFilter = true`)
		b.P(`}`)
		b.P(`if "` + params[0] + `" == "userInfo.phoneNumber" {`)
		b.P(`searchPhone = true`)
		b.P(`}`)
		b.P(`query = query.Must(`, b.elasticPkg.Use(), `.NewMultiMatchQuery(`)
		b.P(variableName + `, "` + params[0] + `.search", "` + params[0] + `.search_reverse").MaxExpansions(1024).Slop(2).Type("phrase_prefix"))`)
		b.P(`} else {`)
		b.P(`fieldsSearch := make([]string, 2*len(fields))`)
		b.P(`for i, field := range fields {`)
		b.P(`fieldsSearch[2*i] = field+ ".search"`)
		b.P(`fieldsSearch[2*i+1] = field+".search_reverse"`)
		b.P(`}`)
		b.P(`query = query.Must(`, b.elasticPkg.Use(), `.NewMultiMatchQuery(`)
		b.P(variableName, `, fieldsSearch...).MaxExpansions(1024).Slop(2).Type("phrase_prefix"))`)
		b.P(`}`)
		b.P(`}`)
	case "*%":
		b.P(`if !checkNull( ` + variableName + `){`)
		b.P(`query = query.Must(`, b.elasticPkg.Use(), `.NewMatchPhrasePrefixQuery(`)
		b.P(`fmt.Sprintf("%s.search", ` + params[0] + `),`)
		b.P(variableName + `,).MaxExpansions(1024).Slop(2))`)
		b.P(`}`)
	case "%*":
		b.P(`if !checkNull( ` + variableName + `){`)
		b.P(`query = query.Must(`, b.elasticPkg.Use(), `.NewMatchPhrasePrefixQuery(`)
		b.P(`fmt.Sprintf("%s.search_reverse", ` + params[0] + `,`)
		b.P(variableName + `,).MaxExpansions(1024).Slop(2))`)
		b.P(`}`)
	case "*.*": //Wildcard
		b.P(`!if checkNull( ` + variableName + `){`)
		b.P(`s := fmt.Sprintf("%v", ` + variableName + `)`)
		b.P(`if !strings.Contains(s, "*") {`)
		b.P(`	s = "*" + s + "*"`)
		b.P(`}`)
		b.P(`query = query.Must(`, b.elasticPkg.Use(), `.NewWildcardQuery(`+params[0]+`, s))`)
		b.P(`}`)
	case "*.": //Wildcard
		b.P(`!if checkNull( ` + variableName + `){`)
		b.P(`query = query.Must(`, b.elasticPkg.Use(), `.NewWildcardQuery(`+params[0]+`, fmt.Sprintf("*%v", `+variableName+`)))`)
		b.P(`}`)
	case ".*": //Wildcard
		b.P(`!if checkNull( ` + variableName + `){`)
		b.P(`query = query.Must(`, b.elasticPkg.Use(), `.NewWildcardQuery(`+params[0]+`, fmt.Sprintf("%v*", `+variableName+`)))`)
		b.P(`}`)
	case "=": //Term'
		b.P(`if !checkNull( ` + variableName + `){`)
		b.P(`if `+b.reflectPkg.Use()+`.TypeOf(`, variableName, `).Kind() == `+b.reflectPkg.Use()+`.Slice {`)
		b.P(`query = query.Filter(`, b.elasticPkg.Use(), `.NewTermsQuery("`+params[0]+`",`+variableName+`))`)
		b.P(`} else if isEnumAll(`, variableName, `) {`)
		b.P(b.glogPkg.Use(), `.Infoln("`+params[0]+`", "Is enum all")`)
		b.P(`} else{`)
		b.P(`comp := convertDateTimeSearch(` + variableName + `,"=")`)
		b.P(`query = query.Filter(`, b.elasticPkg.Use(), `.NewTermQuery("`+params[0]+`",comp))`)
		b.P(`}`)
		b.P(`}`)
	case "mt":
		b.P(`if !checkNull( ` + variableName + `){`)
		b.P(`query = query.Must(`, b.elasticPkg.Use(), `.NewMatchQuery("`+params[0]+`",`+variableName+`))`)
		b.P(`}`)
		// query = query.Must(elastic.NewMatchQuery(params[0], vv))
	case "match":
		b.P(`if !checkNull( ` + variableName + `){`)
		b.P(`query = query.Must(`, b.elasticPkg.Use(), `.NewMatchQuery("`+params[0]+`.search",`+variableName+`).MinimumShouldMatch("3<90%"))`)
		b.P(`}`)
	case ">=":
		b.P(b.glogPkg.Use(), `.Infoln("`+params[0]+`",`+variableName+`)`)
		once.Do(rangeQueryDeclar)
		b.P(`if !checkNull( ` + variableName + `){`)
		b.P(`if !rangeDateSearch.addFrom("` + params[0] + `", ` + variableName + `) {`)
		b.P(`query = query.Must(r.NewRangeQuery("` + params[0] + `").Gte(` + variableName + `))`)
		b.P(`}`)
		b.P(`}`)
	case "<=":
		b.P(b.glogPkg.Use(), `.Infoln("`+params[0]+`",`+variableName+`)`)
		once.Do(rangeQueryDeclar)
		b.P(`if !checkNull( ` + variableName + `){`)
		b.P(`if !rangeDateSearch.addTo("` + params[0] + `", ` + variableName + `) {`)
		b.P(`query = query.Must(r.NewRangeQuery("` + params[0] + `").Lte(` + variableName + `))`)
		b.P(`}`)
		b.P(`}`)
	case ">":
		b.P(b.glogPkg.Use(), `.Infoln("`+params[0]+`",`+variableName+`)`)
		once.Do(rangeQueryDeclar)
		b.P(`if !checkNull( ` + variableName + `){`)
		b.P(`if !rangeDateSearch.addFrom("` + params[0] + `", ` + variableName + `) {`)
		b.P(`query = query.Must(r.NewRangeQuery("` + params[0] + `").Gt(` + variableName + `))`)
		b.P(`}`)
		b.P(`}`)
	case "<":
		b.P(b.glogPkg.Use(), `.Infoln("`+params[0]+`",`+variableName+`)`)
		once.Do(rangeQueryDeclar)
		b.P(`if !checkNull( ` + variableName + `){`)
		b.P(`if !rangeDateSearch.addTo("` + params[0] + `", ` + variableName + `) {`)
		b.P(`	query = query.Must(r.NewRangeQuery("` + params[0] + `").Lt(` + variableName + `))`)
		b.P(`}`)
		b.P(`}`)
	case "!=":
		b.P(b.glogPkg.Use(), `.Infoln("`+params[0]+`",`+variableName+`)`)
		b.P(`if !checkNull( ` + variableName + `){`)
		b.P(`if `, b.reflectPkg.Use(), `.TypeOf(`, variableName, `).Kind() == `, b.reflectPkg.Use(), `.Slice {`)
		b.P(`query = query.MustNot(`, b.elasticPkg.Use(), `.NewTermsQuery("`+params[0]+`",`+variableName+`))`)
		b.P(`} else {`)
		b.P(`comp := convertDateTimeSearch(` + variableName + `,"!=")`)
		b.P(`query = query.MustNot(`, b.elasticPkg.Use(), `.NewTermQuery("`+params[0]+`",comp))`)
		b.P(`}`)
		b.P(`}`)
	default:
		b.P(b.glogPkg.Use(), `.Warningln("Unknow ", params[1])`)
	}

}
