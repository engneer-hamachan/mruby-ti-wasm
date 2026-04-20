package builtin

import (
	"slices"
	"strings"
	"ti/base"
)

type TypeSpec []string

type MethodArgument struct {
	Type       TypeSpec       `json:"type,omitempty"`
	Key        string         `json:"key,omitempty"`
	IsAsterisk bool           `json:"is_asterisk,omitempty"`
	IsDefault  bool           `json:"is_default,omitempty"`
	Value      map[string]any `json:",inline"`
}

type MethodReturn struct {
	Type                TypeSpec `json:"type,omitempty"`
	IsConditionalReturn bool     `json:"is_conditional,omitempty"`
	IsDestructive       bool     `json:"is_destructive,omitempty"`
	IsCaptureOwner      bool     `json:"is_capture_owner,omitempty"`
}

type MethodDefinition struct {
	Name            string           `json:"name"`
	BlockParameters []string         `json:"block_parameters"`
	Arguments       []MethodArgument `json:"arguments"`
	ReturnType      MethodReturn     `json:"return_type"`
	Document        string           `json:"document"`
}

type ConstDefinition struct {
	Name       string       `json:"name"`
	ReturnType MethodReturn `json:"return_type"`
}

type PropertyDefinition struct {
	Name   string   `json:"name"`
	Type   TypeSpec `json:"type"`
	Access string   `json:"access"`
}

type InstanceVarDefinition struct {
	Name string   `json:"name"`
	Type TypeSpec `json:"type"`
}

type ClassDefinition struct {
	Frame              string                  `json:"frame"`
	Class              string                  `json:"class"`
	Type               string                  `json:"type,omitempty"`
	InstanceMethods    []MethodDefinition      `json:"instance_methods"`
	ClassMethods       []MethodDefinition      `json:"class_methods"`
	Constants          []ConstDefinition       `json:"constants"`
	Extends            []string                `json:"extends"`
	InstanceProperties []PropertyDefinition    `json:"instance_properties"`
	InstanceVariables  []InstanceVarDefinition `json:"instance_variables"`
}

var AllTypeNames = []string{
	"NilClass",
	"Symbol",
	"Bool",
	"DefaultBool",
	"Block",
	"DefaultBlock",
	"Range",
	"Untyped",
	"DefaultUntyped",
	"String",
	"DefaultString",
	"OptionalString",
	"Int",
	"DefaultInt",
	"OptionalInt",
	"Float",
	"DefaultFloat",
	"OptionalFloat",
	"Array",
	"Hash",
	"StringArray",
	"IntArray",
	"FloatArray",
	"Self",
	"Number",
	"IntInt",
	"Unify",
	"OptionalUnify",
	"BlockResultArray",
	"SelfArray",
	"Argument",
	"KeyArray",
	"KeyValueArray",
	"UnifyArgument",
	"Flatten",
	"Item",
	"Owner",
	"SymbolToMethod",
	"SymbolToMethods",
}

func parseTypeString(typeStr string) base.T {
	if len(typeStr) > 1 && typeStr[0] == '?' {
		innerType := parseTypeString(typeStr[1:])
		return *base.MakeUnion([]base.T{innerType, NilT})
	}

	if len(typeStr) > 1 && typeStr[0] == '*' {
		innerType := parseTypeString(typeStr[1:])
		innerType.IsBuiltinAsterisk = true
		return innerType
	}

	if len(typeStr) > 2 && typeStr[0] == '[' && typeStr[len(typeStr)-1] == ']' {
		innerType := typeStr[1 : len(typeStr)-1]
		innerT := parseTypeString(innerType)
		return *base.MakeArray([]base.T{innerT})
	}

	if strings.Contains(typeStr, "|") {
		parts := strings.Split(typeStr, "|")
		var types []base.T
		for _, part := range parts {
			types = append(types, parseTypeString(strings.TrimSpace(part)))
		}
		return *base.MakeUnion(types)
	}

	return ConvertToBuiltinT(typeStr)
}

func parseReturnType(returnType MethodReturn) base.T {
	var t base.T

	switch len(returnType.Type) {
	case 0:
		t = NilT
	case 1:
		t = parseTypeString(returnType.Type[0])
	default:
		var types []base.T
		for _, s := range returnType.Type {
			types = append(types, parseTypeString(s))
		}
		t = *base.MakeUnion(types)
	}

	t.IsConditionalReturn = returnType.IsConditionalReturn
	t.IsDestructive = returnType.IsDestructive
	t.IsCaptureOwner = returnType.IsCaptureOwner

	return t
}

func parseArguments(args []MethodArgument) []base.T {
	var result []base.T

	for _, arg := range args {
		var baseType base.T

		switch len(arg.Type) {
		case 0:
			baseType = NilT

		case 1:
			typeStr := arg.Type[0]

			if len(typeStr) > 0 {
				switch typeStr[0] {
				case '*':
					if !strings.Contains(typeStr, "|") && !strings.Contains(typeStr, "[") {
						arg.IsAsterisk = true
						typeStr = typeStr[1:]
						baseType = parseTypeString(typeStr)
					} else {
						baseType = parseTypeString(typeStr)
					}

				case '?':
					if !strings.Contains(typeStr, "|") && !strings.Contains(typeStr, "[") {
						typeStr = typeStr[1:]
						baseType = parseTypeString(typeStr)
						baseType.SetHasDefault(true)
					} else {
						baseType = parseTypeString(typeStr)
					}

				default:
					if base.IsNameSpace(typeStr) {
						frame, parentClass, class := base.SeparateNameSpaces(typeStr)
						baseType = *base.MakeObject(class)
						baseType.SetFrame(base.CalculateFrame(frame, parentClass))
						break
					}
					baseType = parseTypeString(typeStr)
				}
			}

		default:
			var types []base.T
			for _, s := range arg.Type {
				types = append(types, parseTypeString(s))
			}
			baseType = *base.MakeUnion(types)
		}

		baseType.IsBuiltinAsterisk = arg.IsAsterisk
		baseType.SetIsBuiltin(true)

		if arg.IsDefault {
			baseType.SetHasDefault(true)
		}

		switch arg.Key {
		case "":
			result = append(result, baseType)
		default:
			keywordType := base.MakeKeyValue(arg.Key, &baseType)
			result = append(result, *keywordType)
		}
	}

	return result
}

func appendBlockParameters(returnType *base.T, method MethodDefinition) {
	var blockParameters []base.T

	for _, parameter := range method.BlockParameters {
		blockParameters = append(blockParameters, parseTypeString(parameter))
	}

	if len(blockParameters) > 0 {
		returnType.IsBlockGiven = true
	}

	returnType.SetBlockParamaters(blockParameters)
}

func loadClassDef(classDef ClassDefinition) {
	if base.IsNameSpace(classDef.Class) {
		frame, parentClass, class := base.SeparateNameSpaces(classDef.Class)
		t := *base.MakeObject(class)
		t.SetFrame(base.CalculateFrame(frame, parentClass))
		return
	}

	d := NewDefineBuiltinMethod(classDef.Frame, classDef.Class)

	base.BuiltinClasses = append(base.BuiltinClasses, classDef.Class)

	for _, method := range classDef.InstanceMethods {
		args := parseArguments(method.Arguments)
		returnType := parseReturnType(method.ReturnType)

		if len(method.BlockParameters) > 0 {
			appendBlockParameters(&returnType, method)
		}

		key := classDef.Frame + classDef.Class + method.Name
		if method.Document != "" || base.TSignatureDocument[key] == "" {
			base.TSignatureDocument[key] = strings.ReplaceAll(method.Document, "\n", "<CR>")
		}

		d.defineBuiltinInstanceMethod(classDef.Frame, method.Name, args, returnType)
	}

	for _, method := range classDef.ClassMethods {
		args := parseArguments(method.Arguments)
		returnType := parseReturnType(method.ReturnType)

		if len(method.BlockParameters) > 0 {
			appendBlockParameters(&returnType, method)
		}

		key := classDef.Frame + classDef.Class + method.Name + "static"
		if method.Document != "" || base.TSignatureDocument[key] == "" {
			base.TSignatureDocument[key] = strings.ReplaceAll(method.Document, "\n", "<CR>")
		}

		d.defineBuiltinStaticMethod(classDef.Frame, method.Name, args, returnType)
	}

	for _, constant := range classDef.Constants {
		returnType := parseReturnType(constant.ReturnType)
		d.defineBuiltinConstant(classDef.Frame, classDef.Class, constant.Name, returnType)
	}

	for _, prop := range classDef.InstanceProperties {
		propType := parseReturnType(MethodReturn{Type: prop.Type})
		propT := &propType
		if prop.Access == "reader" {
			propT.EnableReadOnly()
		}
		base.SetInstanceValueT(classDef.Frame, classDef.Class, prop.Name, propT)
	}

	for _, ivar := range classDef.InstanceVariables {
		ivarType := parseReturnType(MethodReturn{Type: ivar.Type})
		ivarT := &ivarType
		base.SetInstanceValueT(classDef.Frame, classDef.Class, ivar.Name, ivarT)
	}

	if classDef.Class != "" && classDef.Class != "Kernel" {
		classNode := base.ClassNode{Frame: classDef.Frame, Class: classDef.Class}
		parentNode := base.ClassNode{Frame: "Builtin", Class: ""}
		base.ClassInheritanceMap[classNode.Key()] =
			append(base.ClassInheritanceMap[classNode.Key()], parentNode.Encode())
	}

	for _, class := range classDef.Extends {
		classNode := base.ClassNode{Frame: classDef.Frame, Class: classDef.Class}

		var parentNode base.ClassNode
		parts := strings.Split(class, "::")
		switch len(parts) {
		case 2:
			parentNode = base.ClassNode{Frame: parts[0], Class: parts[1]}
		default:
			parentNode = base.ClassNode{Frame: classDef.Frame, Class: class}
		}

		if slices.Contains(base.ClassInheritanceMap[classNode.Key()], parentNode.Encode()) {
			continue
		}

		base.ClassInheritanceMap[classNode.Key()] =
			append(base.ClassInheritanceMap[classNode.Key()], parentNode.Encode())
	}

	d.SetDefinedClass()
}
