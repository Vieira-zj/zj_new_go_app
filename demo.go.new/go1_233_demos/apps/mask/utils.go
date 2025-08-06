package mask

import "log"

func Output(val any) {
	switch v := val.(type) {
	case string:
		log.Println(v)
	case Sensitive:
		val = v.MaskSensitive()
		log.Printf("%+v", val)
	default:
		log.Println("unexpected type:", v)
	}
}

func MakeSensitive(u any) any {
	if s, ok := u.(Sensitive); ok {
		return s.MaskSensitive()
	}
	return u
}
