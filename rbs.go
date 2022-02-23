package main

import (
	"github.com/fundthmcalculus/protoc-gen-rbi/ruby_types"
	"log"
	"strings"
	"text/template"

	pgs "github.com/lyft/protoc-gen-star"
	pgsgo "github.com/lyft/protoc-gen-star/lang/go"
)

type rbsModule struct {
	*pgs.ModuleBase
	context         pgsgo.Context
	template        *template.Template
	serviceTemplate *template.Template
}

func RBS() *rbsModule { return &rbsModule{ModuleBase: &pgs.ModuleBase{}} }

func (m *rbsModule) InitContext(c pgs.BuildContext) {
	m.ModuleBase.InitContext(c)
	m.context = pgsgo.InitContext(c.Parameters())
	funcs := map[string]interface{}{
		"increment":               m.increment,
		"optional":                m.optional,
		"optionalOneOf":           m.optionalOneOf,
		"willGenerateInvalidRuby": m.willGenerateInvalidRuby,
		"rubyModules":             ruby_types.RubyModules,
		"rubyPackage":             ruby_types.RubyPackage,
		"rubyMessageType":         ruby_types.RubyMessageType,
		"rbsGetterFieldType":      ruby_types.RbsGetterFieldType,
		"rbsSetterFieldType":      ruby_types.RbsSetterFieldType,
		"rbsInitializerFieldType": ruby_types.RbsInitializerFieldType,
		"rubyFieldValue":          ruby_types.RubyFieldValue,
		"rbsMethodParamType":      ruby_types.RbsMethodParamType,
		"rbsMethodReturnType":     ruby_types.RbsMethodReturnType,
	}
	m.template = template.Must(template.New("rbs").Funcs(funcs).Parse(templateRbs))
	m.serviceTemplate = template.Must(template.New("rbsService").Funcs(funcs).Parse(serviceTemplateRbs))
}

func (m *rbsModule) Name() string { return "rbi" }

func (m *rbsModule) Execute(targets map[string]pgs.File, pkgs map[string]pgs.Package) []pgs.Artifact {
	for _, t := range targets {
		m.generate(t)

		grpc, err := m.context.Params().BoolDefault("grpc", true)
		if err != nil {
			log.Panicf("Bad parameter: grpc\n")
		}

		if len(t.Services()) > 0 && grpc {
			m.generateServices(t)
		}
	}
	return m.Artifacts()
}

func (m *rbsModule) generate(f pgs.File) {
	op := strings.TrimSuffix(f.InputPath().String(), ".proto") + "_pb.rbs"
	m.AddGeneratorTemplateFile(op, m.template, f)
}

func (m *rbsModule) generateServices(f pgs.File) {
	op := strings.TrimSuffix(f.InputPath().String(), ".proto") + "_services_pb.rbs"
	m.AddGeneratorTemplateFile(op, m.serviceTemplate, f)
}

func (m *rbsModule) increment(i int) int {
	return i + 1
}

func (m *rbsModule) optional(field pgs.Field) bool {
	return field.Descriptor().GetProto3Optional()
}

func (m *rbsModule) optionalOneOf(oneOf pgs.OneOf) bool {
	return len(oneOf.Fields()) == 1 && oneOf.Fields()[0].Descriptor().GetProto3Optional()
}

func (m *rbsModule) willGenerateInvalidRuby(fields []pgs.Field) bool {
	for _, field := range fields {
		if !validRubyField.MatchString(string(field.Name())) {
			return true
		}
	}
	return false
}

const templateRbs = `# Code generated by protoc-gen-rbi. DO NOT EDIT.
# source: {{ .InputPath }}
{{ range rubyModules . }}
module {{ . }}
end{{ end }}
{{ range .AllMessages }}
class {{ rubyMessageType . }}

  def self.decode: (String) -> {{ rubyMessageType . }}

  def self.encode: ({{ rubyMessageType . }}) -> String

  def self.decode_json: (String, untyped kw) -> {{ rubyMessageType . }}

  def self.encode_json: ({{ rubyMessageType . }}, untyped kw) -> String

  def self.descriptor: () -> Google::Protobuf::Descriptor

  # Constants of the form Constant_1 are invalid. We've declined to type this as a result, taking a hash instead.
  def initialize: (::Hash[untyped, untyped]) -> void

  def initialize: ({{ $index := 0 }}{{ range .Fields }}{{ if gt $index 0 }},{{ end }}{{ $index = increment $index }}
    {{ .Name }}: {{ rbsInitializerFieldType . }}{{ end }}
  ) -> void

{{ range .Fields }}
  def {{ .Name }}: () -> {{ rbsGetterFieldType . }}

  def {{ .Name }}=: ({{ rbsSetterFieldType . }}) -> void

  def clear_{{ .Name }}: () -> void

  def has_{{ .Name }}?: () -> bool
{{ end }}{{ range .OneOfs }}
  def {{ .Name }}: () -> Symbol?
{{ end }}
  def []: (String) -> untyped

  def []=: (String, untyped value) -> void

  def to_h: () -> ::Hash[Symbol, untyped]
end
{{ end }}{{ range .AllEnums }}
module {{ rubyMessageType . }}{{ range .Values }}
  # TODO - Not sure how to represent this
  # self::{{ .Name }} = T.let({{ .Value }}, Integer)
  def {{ .Name }}: () -> Integer # = {{ .Value}}{{ end }}

  def self.lookup: (value: Integer) -> Symbol?

  def self.resolve: (value: Symbol) -> Integer?

  def self.descriptor: () -> ::Google::Protobuf::EnumDescriptor
end
{{ end }}`

const serviceTemplateRbs = `# Code generated by protoc-gen-rbi. DO NOT EDIT.
# source: {{ .InputPath }}
{{ range .Services }}
module {{ rubyPackage .File }}::{{ .Name }}
  class Service
  end

  class Stub < GRPC::ClientStub
    def initialize: (String, GRPC::Core::ChannelCredentials, untyped kw) -> void
    {{ range .Methods }}

    def {{ .Name.LowerSnakeCase }}: ({{ rbsMethodParamType . }}) -> {{ rbsMethodReturnType . }}
    {{ end }}
  end
end
{{ end }}`