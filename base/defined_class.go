package base

var DefinedClassTable = make(map[string]bool)

func definedClassKey(frame, class string) string {
	return frame + "\x00" + class
}

func init() {
	class := "Object"

	classNode := ClassNode{Frame: "Builtin", Class: class}
	objectClassNode := ClassNode{Frame: "Builtin", Class: ""}

	ClassInheritanceMap[classNode.Key()] =
		append(ClassInheritanceMap[classNode.Key()], objectClassNode.Encode())

	returnT := MakeObject(class)
	args := "*" + GenId()
	methodT := MakeMethod("Builtin", "new", *returnT, []string{args})
	SetClassMethodT("", class, methodT, false, "unknown", 0)

	DefinedClassTable[definedClassKey("Builtin", class)] = true
}

func IsClassDefined(frames []string, class string) bool {
	for _, frame := range frames {
		if DefinedClassTable[definedClassKey(frame, class)] {
			return true
		}

		if DefinedClassTable[definedClassKey("Builtin::"+frame, class)] {
			return true
		}
	}

	return DefinedClassTable[definedClassKey("Builtin", class)]
}

func SetDefinedClass(frame, class string) {
	DefinedClassTable[definedClassKey(frame, class)] = true
}
