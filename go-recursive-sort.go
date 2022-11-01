package recursivesort

import (
	"fmt"
	"reflect"
	"sort"
)

func swapValues(slice reflect.Value, tmp reflect.Value, i, j int) {
	vi := slice.Index(i)
	vj := slice.Index(j)
	tmp.Elem().Set(vi)
	vi.Set(vj)
	vj.Set(tmp.Elem())
}

// TypePriorityLookupHelper is an interface to compare two types.
type TypePriorityLookupHelper interface {
	CompareTypes(iv, ij reflect.Value) bool
	CompareUnknownTypes(iv, ij reflect.Value) bool
}

// TypePriorityLookup maps types to integer priorities.
type TypePriorityLookup struct {
	Priorities map[reflect.Type]int
}

// CompareTypes compars two known types based on their priorities.
//
// If priorities are not known, it delegates to `CompareUnknownTypes`.
func (lookup *TypePriorityLookup) CompareTypes(iv, jv reflect.Value) bool {
	ivp, iok := lookup.Priorities[iv.Type()]
	jvp, jok := lookup.Priorities[jv.Type()]
	if !(iok && jok) {
		return lookup.CompareUnknownTypes(iv, jv)
	}
	return ivp < jvp
}

// CompareUnknownTypes compares two types, of which at least one is not known,
// based on their string name.
func (lookup *TypePriorityLookup) CompareUnknownTypes(iv, jv reflect.Value) bool {
	// order based on the type name
	return iv.Type().String() < jv.Type().String()
}

// FromTypes builds a `TypePriorityLookup` priority lookup table
// based on the order of the passed types
func (lookup TypePriorityLookup) FromTypes(order ...reflect.Type) *TypePriorityLookup {
	priorities := make(map[reflect.Type]int)
	for idx, typ := range order {
		priorities[typ] = idx
	}
	return &TypePriorityLookup{
		Priorities: priorities,
	}
}

// FromValues builds a `TypePriorityLookup` priority lookup table
// based on the order of the passed values
func (lookup TypePriorityLookup) FromValues(order ...interface{}) *TypePriorityLookup {
	types := make([]reflect.Type, len(order))
	for idx, i := range order {
		v := reflect.ValueOf(i)
		for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
			v = v.Elem()
		}
		types[idx] = v.Type()
	}
	return lookup.FromTypes(types...)
}

type sortSliceOfInterfaces struct {
	TypePriorityLookupHelper
	v               reflect.Value
	tmp             reflect.Value
	size            int
	structSortField string
	mapSortKey      reflect.Value
	strict          bool
}

func (s sortSliceOfInterfaces) Len() int { return s.size }

func (s sortSliceOfInterfaces) Swap(i, j int) {
	swapValues(s.v, s.tmp, i, j)
}

func getMapKey(v reflect.Value, key reflect.Value) (reflect.Value, bool) {
	for _, mapKey := range v.MapKeys() {
		if mapKey.Interface() == key.Interface() {
			return v.MapIndex(mapKey), true
		}
	}
	return reflect.Value{}, false
}

func (s sortSliceOfInterfaces) compareMap(iv, jv reflect.Value) bool {
	if s.mapSortKey.Kind() == reflect.Invalid {
		// compare by type
		return s.CompareTypes(iv, jv)
	}
	// compare by specific map key if present
	ivKey, iok := getMapKey(iv, s.mapSortKey)
	jvKey, jok := getMapKey(jv, s.mapSortKey)
	if !(iok && jok) {
		if s.strict {
			panic("has no such map key")
		}
		return s.CompareTypes(iv, jv)
	}
	return s.compareValues(ivKey, jvKey)
}

func (s sortSliceOfInterfaces) compareStruct(iv, jv reflect.Value) bool {
	// compare by specific field if present
	fi, fiok := iv.Type().FieldByName(s.structSortField)
	fj, fjok := jv.Type().FieldByName(s.structSortField)
	if !(fiok && fjok) {
		if s.strict {
			panic("no such field")
		}
		if iv.NumField() < 1 || jv.NumField() < 1 {
			// cannot compare empty slices
			panic("cannot compare empty slices")
		}
		// compare values of first exported field
		for i := 0; i < iv.NumField(); i++ {
			fiv := iv.Field(i)
			fjv := jv.Field(i)
			if fiv.CanInterface() && fjv.CanInterface() {
				fivi := reflect.ValueOf(fiv.Interface())
				fjvi := reflect.ValueOf(fjv.Interface())
				return s.compareValues(fivi, fjvi)
			}
		}
		panic("cannot compare struct with no exported fields")
	}

	fiv := iv.Field(fi.Index[0])
	fjv := jv.Field(fj.Index[0])
	fivi := reflect.ValueOf(fiv.Interface())
	fjvi := reflect.ValueOf(fjv.Interface())
	return s.compareValues(fivi, fjvi)
}

func comparePrimitiveInt(iv, jv reflect.Value) bool {
	ivi := iv.Interface()
	jvi := jv.Interface()
	switch ivi.(type) {
	case uint:
		return ivi.(uint) < jvi.(uint)
	case uint8: // also covers byte
		return ivi.(uint8) < jvi.(uint8)
	case uint16:
		return ivi.(uint16) < jvi.(uint16)
	case uint32:
		return ivi.(uint32) < jvi.(uint32)
	case uint64:
		return ivi.(uint64) < jvi.(uint64)

	case int:
		return ivi.(int) < jvi.(int)
	case int8:
		return ivi.(int8) < jvi.(int8)
	case int16:
		return ivi.(int16) < jvi.(int16)
	case int32: // also covers rune
		return ivi.(int32) < jvi.(int32)
	case int64:
		return ivi.(int64) < jvi.(int64)
	default:
		panic(fmt.Sprintf("not implemented: %v", iv.Kind()))
	}
}

func comparePrimitive(iv, jv reflect.Value) bool {
	ivi := iv.Interface()
	jvi := jv.Interface()
	switch ivi.(type) {
	// bool
	case bool:
		if ivi.(bool) == false && jvi.(bool) == true {
			return true
		}
		return false

		// string
	case string:
		return ivi.(string) < jvi.(string)

		// integers
	case uint, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64:
		return comparePrimitiveInt(iv, jv)

		// float numbers
	case float32:
		return ivi.(float32) < jvi.(float32)
	case float64:
		return ivi.(float64) < jvi.(float64)

		// complex numbers
		// note: there is no total ordering of complex numbers
		// we resort to lexicographic ordering
	case complex64:
		i := ivi.(complex64)
		j := jvi.(complex64)
		if real(i) < real(j) {
			return true
		}
		return imag(i) < imag(j)
	case complex128:
		i := ivi.(complex128)
		j := jvi.(complex128)
		if real(i) < real(j) {
			return true
		}
		return imag(i) < imag(j)
	default:
	}
	panic(fmt.Sprintf("not implemented: %v", iv.Kind()))
}

func (s sortSliceOfInterfaces) compareSameKind(iv, jv reflect.Value) bool {
	switch iv.Kind() {
	case reflect.Map:
		return s.compareMap(iv, jv)
	case reflect.Struct:
		return s.compareStruct(iv, jv)
	default:
		return comparePrimitive(iv, jv)
	}
}

func (s sortSliceOfInterfaces) compareValues(iv, jv reflect.Value) bool {
	// Indirect through pointers and interfaces
	for iv.Kind() == reflect.Ptr || iv.Kind() == reflect.Interface {
		iv = iv.Elem()
	}
	for jv.Kind() == reflect.Ptr || jv.Kind() == reflect.Interface {
		jv = jv.Elem()
	}
	// compare directly if of same type
	if iv.Kind() == jv.Kind() {
		return s.compareSameKind(iv, jv)
	}
	// otherwise sort based on type priority
	return s.CompareTypes(iv, jv)
}

func (s sortSliceOfInterfaces) Less(i, j int) bool {
	iv, jv := s.v.Index(i), s.v.Index(j)
	return s.compareValues(iv, jv)
}

// RecursiveSort implements a recursive sort interface for arbitrary types.
type RecursiveSort struct {
	TypePriorityLookupHelper

	// MapSortKey specifies the key of maps to use as the sorting key if available
	MapSortKey interface{}

	// StructSortField specifies the field of structs to use as the sorting key if available
	StructSortField string

	// Strict forces using `StructSortField` and `MapSortKey`.
	//
	// Note: if the key or field does not exist, this will panic.
	Strict bool
}

func (rec *RecursiveSort) sortMap(v reflect.Value) {
	for _, k := range v.MapKeys() {
		rec.sort(v.MapIndex(k))
	}
}

// Sort recursively sorts an interface.
func Sort(v interface{}) {
	sort := &RecursiveSort{}
	sort.Sort(v)
}

// Sort recursively sorts an interface.
func (rec *RecursiveSort) Sort(v interface{}) {
	rec.sort(reflect.ValueOf(v))
}

func (rec *RecursiveSort) sort(v reflect.Value) {
	if rec.TypePriorityLookupHelper == nil {
		rec.TypePriorityLookupHelper = TypePriorityLookup{}.FromTypes()
	}
	if !v.CanInterface() {
		// not exported, skip
		return
	}
	// Indirect through pointers and interfaces
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		// sort slice elements first
		for i := 0; i < v.Len(); i++ {
			rec.sort(v.Index(i))
		}
		sortFunc := sortSliceOfInterfaces{
			v:                        v,
			tmp:                      reflect.New(v.Type().Elem()),
			size:                     v.Len(),
			mapSortKey:               reflect.ValueOf(rec.MapSortKey),
			structSortField:          rec.StructSortField,
			strict:                   rec.Strict,
			TypePriorityLookupHelper: rec.TypePriorityLookupHelper,
		}
		sort.Sort(sortFunc)
	case reflect.Map:
		for _, k := range v.MapKeys() {
			rec.sort(v.MapIndex(k))
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			rec.sort(field)
		}
	default:
		// ignore for now
	}
}
