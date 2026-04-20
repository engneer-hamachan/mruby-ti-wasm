package base

type ClassNode struct {
	Frame     string
	Class     string
	IsInclude bool
	IsExtend  bool
}

// Key returns the map key for ClassInheritanceMap lookups.
// Keys never have IsInclude/IsExtend set.
func (c ClassNode) Key() string {
	return c.Frame + "\x00" + c.Class
}

// Encode returns a string encoding of this ClassNode for storage as a map value.
func (c ClassNode) Encode() string {
	ie := byte('0')
	if c.IsInclude {
		ie = '1'
	}
	ext := byte('0')
	if c.IsExtend {
		ext = '1'
	}
	return string([]byte{ie, ext}) + "\x00" + c.Frame + "\x00" + c.Class
}

// DecodeClassNode decodes a ClassNode from the Encode() format.
func DecodeClassNode(s string) ClassNode {
	if len(s) < 3 {
		return ClassNode{}
	}
	ie := s[0] == '1'
	ext := s[1] == '1'
	rest := s[3:] // skip "ie\x00"
	i := 0
	for i < len(rest) && rest[i] != 0 {
		i++
	}
	frame := rest[:i]
	class := ""
	if i+1 < len(rest) {
		class = rest[i+1:]
	}
	return ClassNode{
		Frame:     frame,
		Class:     class,
		IsInclude: ie,
		IsExtend:  ext,
	}
}

func ParseClassNodeKey(key string) ClassNode {
	for i, b := range key {
		if b == '\x00' {
			return ClassNode{Frame: key[:i], Class: key[i+1:]}
		}
	}
	return ClassNode{Frame: key}
}

// ClassInheritanceMap stores parent ClassNodes encoded as strings to avoid
// TinyGo GC issues with complex struct-containing slice values in maps.
var ClassInheritanceMap = make(map[string][]string)
