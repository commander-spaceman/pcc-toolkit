package pccenc

import (
	"errors"
	"fmt"
	"math"
)

type PropertyValue struct {
	Name       string
	PropType   string
	Value      interface{}
	ArrayIndex int

	StructTypeName   string
	ByteSubTypeName  string
	ArrayElementType string
	Properties       []PropertyValue
	Items            []PropertyValue
}

const (
	headerBaseSize = 24
	fnameByteSize  = 8
)

type encodeCtx struct {
	names      []string
	nameLookup map[string]int
}

func EncodePropertyValue(pv PropertyValue, names []string) ([]byte, error) {
	ctx := &encodeCtx{
		names:      names,
		nameLookup: buildNameLookup(names),
	}
	return ctx.encodeProperty(pv)
}

func EncodeNoneProperty(names []string) ([]byte, error) {
	ctx := &encodeCtx{
		names:      names,
		nameLookup: buildNameLookup(names),
	}
	return ctx.encodeNoneProperty()
}

func EncodePropertyCollection(props []PropertyValue, names []string) ([]byte, error) {
	ctx := &encodeCtx{
		names:      names,
		nameLookup: buildNameLookup(names),
	}
	return ctx.encodeCollection(props)
}

func (ctx *encodeCtx) encodeProperty(pv PropertyValue) ([]byte, error) {
	nameIdx, err := resolveNameIdx(pv.Name, ctx.names, ctx.nameLookup)
	if err != nil {
		return nil, err
	}
	typeIdx, err := resolveTypeIdx(pv.PropType, ctx.names, ctx.nameLookup)
	if err != nil {
		return nil, err
	}

	meta, metaSize, err := ctx.encodeMetadata(pv)
	if err != nil {
		return nil, err
	}

	valueBytes, err := ctx.encodeValue(pv)
	if err != nil {
		return nil, err
	}

	totalSize := headerBaseSize + metaSize + len(valueBytes)
	w := newWriteBuffer(totalSize)

	w.WriteFName(nameIdx, 0)
	w.WriteFName(typeIdx, 0)
	w.WriteI32(len(valueBytes))
	w.WriteI32(pv.ArrayIndex)

	if meta != nil {
		w.WriteBytes(meta)
	}
	if len(valueBytes) > 0 {
		w.WriteBytes(valueBytes)
	}

	return w.Bytes(), nil
}

func (ctx *encodeCtx) encodeMetadata(pv PropertyValue) ([]byte, int, error) {
	switch pv.PropType {
	case "StructProperty":
		if pv.StructTypeName == "" {
			return nil, 0, errors.New("StructProperty requires StructTypeName")
		}
		idx, err := resolveNameIdx(pv.StructTypeName, ctx.names, ctx.nameLookup)
		if err != nil {
			return nil, 0, fmt.Errorf("struct type %q: %w", pv.StructTypeName, err)
		}
		mw := newWriteBuffer(fnameByteSize)
		mw.WriteFName(idx, 0)
		return mw.Bytes(), fnameByteSize, nil

	case "ByteProperty":
		if pv.ByteSubTypeName == "" {
			return nil, 0, errors.New("ByteProperty requires ByteSubTypeName")
		}
		idx, err := resolveNameIdx(pv.ByteSubTypeName, ctx.names, ctx.nameLookup)
		if err != nil {
			return nil, 0, fmt.Errorf("byte subtype %q: %w", pv.ByteSubTypeName, err)
		}
		mw := newWriteBuffer(fnameByteSize)
		mw.WriteFName(idx, 0)
		return mw.Bytes(), fnameByteSize, nil

	case "BoolProperty":
		mw := newWriteBuffer(4)
		val := boolToInt(pv.Value)
		mw.WriteI32(val)
		return mw.Bytes(), 4, nil

	case "ArrayProperty":
		if pv.ArrayElementType == "" {
			return nil, 0, errors.New("ArrayProperty requires ArrayElementType")
		}
		idx, err := resolveNameIdx(pv.ArrayElementType, ctx.names, ctx.nameLookup)
		if err != nil {
			return nil, 0, fmt.Errorf("array element type %q: %w", pv.ArrayElementType, err)
		}
		mw := newWriteBuffer(fnameByteSize)
		mw.WriteFName(idx, 0)
		return mw.Bytes(), fnameByteSize, nil

	default:
		return nil, 0, nil
	}
}

func (ctx *encodeCtx) encodeValue(pv PropertyValue) ([]byte, error) {
	switch pv.PropType {
	case "IntProperty", "ObjectProperty", "StringRefProperty":
		return encodeInt32Value(pv.Value)

	case "FloatProperty":
		return encodeFloat32Value(pv.Value)

	case "NameProperty", "EnumProperty":
		return ctx.encodeNameValue(pv.Value)

	case "StrProperty":
		return encodeStrValue(pv.Value)

	case "BoolProperty":
		return nil, nil

	case "StructProperty":
		return ctx.encodeCollection(pv.Properties)

	case "ByteProperty":
		return encodeByteValue(pv.Value)

	case "ArrayProperty":
		return ctx.encodeArrayValue(pv)
	}

	return nil, nil
}

func encodeInt32Value(v interface{}) ([]byte, error) {
	ival, ok := toInt(v)
	if !ok {
		return nil, fmt.Errorf("expected int value, got %T", v)
	}
	w := newWriteBuffer(4)
	w.WriteI32(ival)
	return w.Bytes(), nil
}

func encodeFloat32Value(v interface{}) ([]byte, error) {
	fval, ok := toFloat32(v)
	if !ok {
		return nil, fmt.Errorf("expected float32 value, got %T", v)
	}
	w := newWriteBuffer(4)
	bits := math.Float32bits(fval)
	w.WriteU32(bits)
	return w.Bytes(), nil
}

func (ctx *encodeCtx) encodeNameValue(v interface{}) ([]byte, error) {
	name, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("expected string for NameProperty/EnumProperty, got %T", v)
	}
	idx, err := resolveNameIdx(name, ctx.names, ctx.nameLookup)
	if err != nil {
		return nil, err
	}
	w := newWriteBuffer(fnameByteSize)
	w.WriteFName(idx, 0)
	return w.Bytes(), nil
}

func encodeStrValue(v interface{}) ([]byte, error) {
	s, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("expected string for StrProperty, got %T", v)
	}
	return writeUnrealStringLatin1(s), nil
}

func (ctx *encodeCtx) encodeArrayValue(pv PropertyValue) ([]byte, error) {
	if len(pv.Items) == 0 {
		w := newWriteBuffer(4)
		w.WriteI32(0)
		return w.Bytes(), nil
	}

	var itemBytes [][]byte
	totalItemLen := 0
	for _, item := range pv.Items {
		encoded, err := ctx.encodeArrayItem(pv.ArrayElementType, item)
		if err != nil {
			return nil, err
		}
		itemBytes = append(itemBytes, encoded)
		totalItemLen += len(encoded)
	}

	w := newWriteBuffer(4 + totalItemLen)
	w.WriteI32(len(pv.Items))
	for _, ib := range itemBytes {
		w.WriteBytes(ib)
	}
	return w.Bytes(), nil
}

func (ctx *encodeCtx) encodeArrayItem(elemType string, pv PropertyValue) ([]byte, error) {
	if elemType == "StructProperty" {
		props, err := ctx.encodeCollection(pv.Properties)
		if err != nil {
			return nil, err
		}
		return props, nil
	}
	if elemType == "IntProperty" {
		return encodeInt32Value(pv.Value)
	}
	if elemType == "NameProperty" {
		return ctx.encodeNameValue(pv.Value)
	}
	return nil, fmt.Errorf("unsupported array element type: %s", elemType)
}

func (ctx *encodeCtx) encodeCollection(props []PropertyValue) ([]byte, error) {
	var chunks [][]byte
	total := 0
	for _, pv := range props {
		enc, err := ctx.encodeProperty(pv)
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, enc)
		total += len(enc)
	}
	none, err := ctx.encodeNoneProperty()
	if err != nil {
		return nil, err
	}
	chunks = append(chunks, none)
	total += len(none)

	w := newWriteBuffer(total)
	for _, c := range chunks {
		w.WriteBytes(c)
	}
	return w.Bytes(), nil
}

func (ctx *encodeCtx) encodeNoneProperty() ([]byte, error) {
	idx, err := resolveNameIdx("None", ctx.names, ctx.nameLookup)
	if err != nil {
		return nil, err
	}
	w := newWriteBuffer(fnameByteSize)
	w.WriteFName(idx, 0)
	return w.Bytes(), nil
}

func encodeByteValue(v interface{}) ([]byte, error) {
	switch val := v.(type) {
	case int:
		w := newWriteBuffer(4)
		w.WriteI32(val)
		return w.Bytes(), nil
	case string:
		return nil, errors.New("string ByteProperty value not yet supported")
	default:
		return nil, fmt.Errorf("unsupported ByteProperty value type: %T", v)
	}
}

func boolToInt(v interface{}) int {
	b, ok := v.(bool)
	if !ok {
		return 0
	}
	if b {
		return 1
	}
	return 0
}

func toInt(v interface{}) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int32:
		return int(val), true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	case *int:
		if val != nil {
			return *val, true
		}
	}
	return 0, false
}

func toFloat32(v interface{}) (float32, bool) {
	switch val := v.(type) {
	case float32:
		return val, true
	case float64:
		return float32(val), true
	}
	return 0, false
}
