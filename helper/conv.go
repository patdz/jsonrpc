package helper

import "encoding/json"

func Interface2Bool(v interface{}) (rv bool, ok bool) {
	if v == nil {
		return
	}
	rv, ok = v.(bool)
	return
}

func Interface2Uint64(v interface{}) (rv uint64, ok bool) {
	if v == nil {
		return
	}
	fv, ok := v.(float64)
	if !ok {
		return
	}
	rv = uint64(fv)
	return
}

func Interface2Int64(v interface{}) (rv int64, ok bool) {
	if v == nil {
		return
	}
	fv, ok := v.(float64)
	if !ok {
		return
	}
	rv = int64(fv)
	return
}

func Interface2String(v interface{}) (rv string, ok bool) {
	if v == nil {
		return
	}
	rv, ok = v.(string)
	return
}

func Interface2JsonBytes(v interface{}) (rv []byte, ok bool) {
	if v == nil {
		return
	}
	rv, err := json.Marshal(v)
	if err != nil {
		ok = false
		return
	}
	return
}

func Interface2Vector(v interface{}) (rv []interface{}, ok bool) {
	if v == nil {
		return
	}
	rv, ok = v.([]interface{})
	return
}

func Interface2StringVector(v interface{}) (rv []string, ok bool) {
	if v == nil {
		return
	}
	sv, ok := v.([]interface{})
	if !ok {
		return
	}

	var s string
	for _, sr := range sv {
		s, ok = Interface2String(sr)
		if !ok {
			return
		}
		rv = append(rv, s)
	}
	return
}
