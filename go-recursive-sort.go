package gorecursivesort

import (
	"encoding/json"
	"fmt"
	"reflect"
	sortimpl "sort"

	"github.com/google/go-cmp/cmp"
)

// Version is incremented using bump2version
const Version = "0.0.1"

func swapValues(slice reflect.Value, tmp reflect.Value, i, j int) {
	vi := slice.Index(i)
	vj := slice.Index(j)
	tmp.Elem().Set(vi)
	vi.Set(vj)
	vj.Set(tmp.Elem())
}

// TypePriorityLookupHelper ...
type TypePriorityLookupHelper interface {
	CompareTypes(iv, ij reflect.Value) bool
	CompareUnknownTypes(iv, ij reflect.Value) bool
}

// TypePriorityLookup ...
type TypePriorityLookup struct {
	Priorities map[reflect.Type]int
}

// CompareTypes ...
func (lookup *TypePriorityLookup) CompareTypes(iv, jv reflect.Value) bool {
	ivp, iok := lookup.Priorities[iv.Type()]
	jvp, jok := lookup.Priorities[jv.Type()]
	if !(iok && jok) {
		return lookup.CompareUnknownTypes(iv, jv)
	}
	return ivp < jvp
}

// CompareUnknownTypes ...
func (lookup *TypePriorityLookup) CompareUnknownTypes(iv, jv reflect.Value) bool {
	// order based on the type name
	return iv.Type().String() < jv.Type().String()
}

// FromTypes ...
func (lookup TypePriorityLookup) FromTypes(order ...reflect.Type) *TypePriorityLookup {
	priorities := make(map[reflect.Type]int)
	for idx, typ := range order {
		priorities[typ] = idx
	}
	return &TypePriorityLookup{
		Priorities: priorities,
	}
}

// FromValues ...
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
		// just use field zero
		fiv := iv.Field(0)
		fjv := jv.Field(0)
		return s.compareValues(reflect.ValueOf(fiv.Interface()), reflect.ValueOf(fjv.Interface()))
	}
	fiv := iv.Field(fi.Index[0])
	fjv := jv.Field(fj.Index[0])
	return s.compareValues(reflect.ValueOf(fiv.Interface()), reflect.ValueOf(fjv.Interface()))
}

func (s sortSliceOfInterfaces) compareSameKind(iv, jv reflect.Value) bool {
	switch iv.Kind() {
	case reflect.String:
		return iv.String() < jv.String()
	case reflect.Int:
		return iv.Int() < jv.Int()
	case reflect.Map:
		return s.compareMap(iv, jv)
	case reflect.Struct:
		return s.compareStruct(iv, jv)
	default:
		panic(fmt.Sprintf("not implemented: %v", iv.Kind()))
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

// RecursiveSort ...
type RecursiveSort struct {
	TypePriorityLookupHelper
	MapSortKey      interface{}
	StructSortField string
	Strict          bool
}

func (rec *RecursiveSort) sortMap(v reflect.Value) {
	for _, k := range v.MapKeys() {
		rec.sort(v.MapIndex(k))
	}
}

// Sort ...
func (rec *RecursiveSort) Sort(v interface{}) {
	rec.sort(reflect.ValueOf(v))
}

func (rec *RecursiveSort) sort(v reflect.Value) {
	if rec.TypePriorityLookupHelper == nil {
		rec.TypePriorityLookupHelper = TypePriorityLookup{}.FromTypes()
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
		sortimpl.Sort(sortFunc)
	case reflect.Map:
		for _, k := range v.MapKeys() {
			rec.sort(v.MapIndex(k))
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			rec.sort(v.Field(i))
		}
	default:
		// ignore for now
	}
}

// AreEqualJSON ...
func AreEqualJSON(s1, s2 string) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(s1), &o1)
	if err != nil {
		return false, fmt.Errorf("Error mashalling first: %v", err)
	}
	recsort := RecursiveSort{}
	recsort.Sort(o1)
	err = json.Unmarshal([]byte(s2), &o2)
	if err != nil {
		return false, fmt.Errorf("Error mashalling second: %v", err)
	}
	recsort.Sort(o2)
	if diff := cmp.Diff(o1, o2); diff != "" {
		return false, fmt.Errorf("difference: %s", diff)
	}
	return true, nil
}
