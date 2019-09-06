package buildquery

import (
	"fmt"
	"os"

	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	querier "github.com/tvducmt/go-proto-buildquery"
)

type buildquery struct {
	*generator.Generator
	generator.PluginImports
	querierPkg generator.Single
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
	proto3 := gogoproto.IsProto3(file.FileDescriptorProto)
	b.PluginImports = generator.NewPluginImports(b.Generator)

	fmtPkg := b.NewImport("fmt")
	stringsPkg := b.NewImport("strings")
	protoPkg := b.NewImport("github.com/gogo/protobuf/proto")
	if !gogoproto.ImportsGoGoProto(file.FileDescriptorProto) {
		protoPkg = b.NewImport("github.com/golang/protobuf/proto")
	}
	// sortPkg := b.NewImport("sort")
	// strconvPkg := b.NewImport("strconv")
	// reflectPkg := b.NewImport("reflect")
	// sortKeysPkg := b.NewImport("github.com/gogo/protobuf/sortkeys")
	b.querierPkg = b.NewImport("github.com/tvducmt/go-proto-buildquery")

	extensionToGoStringUsed := false

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

func getOneofQueryIfAny(oneof *descriptor.OneofDescriptorProto) *querier.OneofValidator {

}
func (b *buildquery) generateProto3Message(file *generator.FileDescriptor, message *generator.Descriptor) {
	ccTypeName := generator.CamelCaseSlice(message.TypeName())
	b.P(`func (this *`, ccTypeName, `) Validate() error {`)
	b.In()

	for _, oneof := range message.OneofDecl {
		oneofValidator := getOneofQueryIfAny(oneof)
		if oneofValidator == nil {
			continue
		}
		if oneofValidator.GetRequired() {
			oneOfName := generator.CamelCase(oneof.GetName())
			b.P(`if this.Get` + oneOfName + `() == nil {`)
			b.In()
			b.P(`return `, b.validatorPkg.Use(), `.FieldError("`, oneOfName, `",`, b.fmtPkg.Use(), `.Errorf("one of the fields must be set"))`)
			b.Out()
			b.P(`}`)
		}
	}
	for _, field := range message.Field {
		fieldValidator := getFieldValidatorIfAny(field)
		if fieldValidator == nil && !field.IsMessage() {
			continue
		}
		isOneOf := field.OneofIndex != nil
		fieldName := b.GetOneOfFieldName(message, field)
		variableName := "this." + fieldName
		repeated := field.IsRepeated()
		// Golang's proto3 has no concept of unset primitive fields
		nullable := (gogoproto.IsNullable(field) || !gogoproto.ImportsGoGoProto(file.FileDescriptorProto)) && field.IsMessage()
		if b.fieldIsProto3Map(file, message, field) {
			b.P(`// Validation of proto3 map<> fields is unsupported.`)
			continue
		}
		if isOneOf {
			b.In()
			oneOfName := b.GetFieldName(message, field)
			oneOfType := b.OneOfTypeName(message, field)
			//if x, ok := m.GetType().(*OneOfMessage3_OneInt); ok {
			b.P(`if oneOfNester, ok := this.Get` + oneOfName + `().(* ` + oneOfType + `); ok {`)
			variableName = "oneOfNester." + b.GetOneOfFieldName(message, field)
		}
		if repeated {
			b.generateRepeatedCountValidator(variableName, ccTypeName, fieldName, fieldValidator)
			if field.IsMessage() || b.validatorWithNonRepeatedConstraint(fieldValidator) {
				b.P(`for _, item := range `, variableName, `{`)
				b.In()
				variableName = "item"
			}
		} else if fieldValidator != nil {
			if fieldValidator.RepeatedCountMin != nil {
				fmt.Fprintf(os.Stderr, "WARNING: field %v.%v is not repeated, validator.min_elts has no effects\n", ccTypeName, fieldName)
			}
			if fieldValidator.RepeatedCountMax != nil {
				fmt.Fprintf(os.Stderr, "WARNING: field %v.%v is not repeated, validator.max_elts has no effects\n", ccTypeName, fieldName)
			}
		}
		if field.IsString() {
			b.generateStringValidator(variableName, ccTypeName, fieldName, fieldValidator)
		} else if b.isSupportedInt(field) {
			b.generateIntValidator(variableName, ccTypeName, fieldName, fieldValidator)
		} else if field.IsEnum() {
			b.generateEnumValidator(field, variableName, ccTypeName, fieldName, fieldValidator)
		} else if b.isSupportedFloat(field) {
			b.generateFloatValidator(variableName, ccTypeName, fieldName, fieldValidator)
		} else if field.IsBytes() {
			b.generateLengthValidator(variableName, ccTypeName, fieldName, fieldValidator)
		} else if field.IsMessage() {
			if b.validatorWithMessageExists(fieldValidator) {
				if nullable && !repeated {
					b.P(`if nil == `, variableName, `{`)
					b.In()
					b.P(`return `, b.validatorPkg.Use(), `.FieldError("`, fieldName, `",`, b.fmtPkg.Use(), `.Errorf("message must exist"))`)
					b.Out()
					b.P(`}`)
				} else if repeated {
					fmt.Fprintf(os.Stderr, "WARNING: field %v.%v is repeated, validator.msg_exists has no effect\n", ccTypeName, fieldName)
				} else if !nullable {
					fmt.Fprintf(os.Stderr, "WARNING: field %v.%v is a nullable=false, validator.msg_exists has no effect\n", ccTypeName, fieldName)
				}
			}
			if nullable {
				b.P(`if `, variableName, ` != nil {`)
				b.In()
			} else {
				// non-nullable fields in proto3 store actual structs, we need pointers to operate on interfaces
				variableName = "&(" + variableName + ")"
			}
			b.P(`if err := `, b.validatorPkg.Use(), `.CallValidatorIfExists(`, variableName, `); err != nil {`)
			b.In()
			b.P(`return `, b.validatorPkg.Use(), `.FieldError("`, fieldName, `", err)`)
			b.Out()
			b.P(`}`)
			if nullable {
				b.Out()
				b.P(`}`)
			}
		}
		if repeated && (field.IsMessage() || b.validatorWithNonRepeatedConstraint(fieldValidator)) {
			// end the repeated loop
			b.Out()
			b.P(`}`)
		}
		if isOneOf {
			// end the oneof if statement
			b.Out()
			b.P(`}`)
		}
	}
	b.P(`return nil`)
	b.Out()
	b.P(`}`)
}
