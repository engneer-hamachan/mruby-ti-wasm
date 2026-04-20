package base

// FrameKey encodes a TFrame lookup key.
// It is used internally to build string keys for map[string] lookups,
// avoiding TinyGo GC issues with struct-keyed maps.
type FrameKey struct {
	frame          string
	targetClass    string
	targetMethod   string
	targetVariable string
	isPrivate      bool
	isStatic       bool
}

func (f FrameKey) Variable() string {
	return f.targetVariable
}

// Key returns the string encoding of this FrameKey.
func (f FrameKey) Key() string {
	ip := byte('0')
	if f.isPrivate {
		ip = '1'
	}
	is := byte('0')
	if f.isStatic {
		is = '1'
	}
	return f.frame + "\x00" + f.targetClass + "\x00" + f.targetMethod + "\x00" + f.targetVariable + "\x00" + string([]byte{ip, is})
}

// ParseFrameKey decodes a string key produced by FrameKey.Key().
func ParseFrameKey(key string) FrameKey {
	var parts [5]string
	idx := 0
	start := 0
	for i := 0; i < len(key) && idx < 4; i++ {
		if key[i] == 0 {
			parts[idx] = key[start:i]
			idx++
			start = i + 1
		}
	}
	parts[idx] = key[start:]

	fk := FrameKey{
		frame:          parts[0],
		targetClass:    parts[1],
		targetMethod:   parts[2],
		targetVariable: parts[3],
	}
	if len(parts[4]) >= 2 {
		fk.isPrivate = parts[4][0] == '1'
		fk.isStatic = parts[4][1] == '1'
	}
	return fk
}

func classMethodTFrameKey(
	frame string,
	targetClass string,
	targetMethod string,
	isPrivate bool,
) string {
	return FrameKey{
		frame:        frame,
		targetClass:  targetClass,
		targetMethod: targetMethod,
		isPrivate:    isPrivate,
		isStatic:     true,
	}.Key()
}

func methodTFrameKey(
	frame string,
	targetClass string,
	targetMethod string,
	isPrivate bool,
) string {
	return FrameKey{
		frame:        frame,
		targetClass:  targetClass,
		targetMethod: targetMethod,
		isPrivate:    isPrivate,
	}.Key()
}

func valueTFrameKey(
	frame string,
	targetClass string,
	targetMethod string,
	targetVariable string,
	isStatic bool,
) string {
	return FrameKey{
		frame:          frame,
		targetClass:    targetClass,
		targetMethod:   targetMethod,
		targetVariable: targetVariable,
		isStatic:       isStatic,
	}.Key()
}

func constTFrameKey(
	frame string,
	targetVariable string,
) string {
	return FrameKey{
		frame:          frame,
		targetVariable: targetVariable,
	}.Key()
}
