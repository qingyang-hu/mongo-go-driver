// Copyright (C) MongoDB, Inc. 2017-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package bson

import (
	"fmt"
	"math"
	"reflect"

	"go.mongodb.org/mongo-driver/bson/bsonoptions"
)

// UIntCodec is the Codec used for uint values.
//
// Deprecated: Use [go.mongodb.org/mongo-driver/bson.NewRegistry] to get a registry with the
// UIntCodec registered.
type UIntCodec struct {
	// EncodeToMinSize causes EncodeValue to marshal Go uint values (excluding uint64) as the
	// minimum BSON int size (either 32-bit or 64-bit) that can represent the integer value.
	//
	// Deprecated: Use bson.Encoder.IntMinSize instead.
	EncodeToMinSize bool
}

var (
	defaultUIntCodec = NewUIntCodec()

	// Assert that defaultUIntCodec satisfies the typeDecoder interface, which allows it to be used
	// by collection type decoders (e.g. map, slice, etc) to set individual values in a collection.
	_ typeDecoder = defaultUIntCodec
)

// NewUIntCodec returns a UIntCodec with options opts.
//
// Deprecated: Use [go.mongodb.org/mongo-driver/bson.NewRegistry] to get a registry with the
// UIntCodec registered.
func NewUIntCodec(opts ...*bsonoptions.UIntCodecOptions) *UIntCodec {
	uintOpt := bsonoptions.MergeUIntCodecOptions(opts...)

	codec := UIntCodec{}
	if uintOpt.EncodeToMinSize != nil {
		codec.EncodeToMinSize = *uintOpt.EncodeToMinSize
	}
	return &codec
}

// EncodeValue is the ValueEncoder for uint types.
func (uic *UIntCodec) EncodeValue(ec EncodeContext, vw ValueWriter, val reflect.Value) error {
	switch val.Kind() {
	case reflect.Uint8, reflect.Uint16:
		return vw.WriteInt32(int32(val.Uint()))
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		u64 := val.Uint()

		// If ec.MinSize or if encodeToMinSize is true for a non-uint64 value we should write val as an int32
		useMinSize := ec.MinSize || (uic.EncodeToMinSize && val.Kind() != reflect.Uint64)

		if u64 <= math.MaxInt32 && useMinSize {
			return vw.WriteInt32(int32(u64))
		}
		if u64 > math.MaxInt64 {
			return fmt.Errorf("%d overflows int64", u64)
		}
		return vw.WriteInt64(int64(u64))
	}

	return ValueEncoderError{
		Name:     "UintEncodeValue",
		Kinds:    []reflect.Kind{reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint},
		Received: val,
	}
}

func (uic *UIntCodec) decodeType(dc DecodeContext, vr ValueReader, t reflect.Type) (reflect.Value, error) {
	var i64 int64
	var err error
	switch vrType := vr.Type(); vrType {
	case TypeInt32:
		i32, err := vr.ReadInt32()
		if err != nil {
			return emptyValue, err
		}
		i64 = int64(i32)
	case TypeInt64:
		i64, err = vr.ReadInt64()
		if err != nil {
			return emptyValue, err
		}
	case TypeDouble:
		f64, err := vr.ReadDouble()
		if err != nil {
			return emptyValue, err
		}
		if !dc.Truncate && math.Floor(f64) != f64 {
			return emptyValue, errCannotTruncate
		}
		if f64 > float64(math.MaxInt64) {
			return emptyValue, fmt.Errorf("%g overflows int64", f64)
		}
		i64 = int64(f64)
	case TypeBoolean:
		b, err := vr.ReadBoolean()
		if err != nil {
			return emptyValue, err
		}
		if b {
			i64 = 1
		}
	case TypeNull:
		if err = vr.ReadNull(); err != nil {
			return emptyValue, err
		}
	case TypeUndefined:
		if err = vr.ReadUndefined(); err != nil {
			return emptyValue, err
		}
	default:
		return emptyValue, fmt.Errorf("cannot decode %v into an integer type", vrType)
	}

	switch t.Kind() {
	case reflect.Uint8:
		if i64 < 0 || i64 > math.MaxUint8 {
			return emptyValue, fmt.Errorf("%d overflows uint8", i64)
		}

		return reflect.ValueOf(uint8(i64)), nil
	case reflect.Uint16:
		if i64 < 0 || i64 > math.MaxUint16 {
			return emptyValue, fmt.Errorf("%d overflows uint16", i64)
		}

		return reflect.ValueOf(uint16(i64)), nil
	case reflect.Uint32:
		if i64 < 0 || i64 > math.MaxUint32 {
			return emptyValue, fmt.Errorf("%d overflows uint32", i64)
		}

		return reflect.ValueOf(uint32(i64)), nil
	case reflect.Uint64:
		if i64 < 0 {
			return emptyValue, fmt.Errorf("%d overflows uint64", i64)
		}

		return reflect.ValueOf(uint64(i64)), nil
	case reflect.Uint:
		if i64 < 0 || int64(uint(i64)) != i64 { // Can we fit this inside of an uint
			return emptyValue, fmt.Errorf("%d overflows uint", i64)
		}

		return reflect.ValueOf(uint(i64)), nil
	default:
		return emptyValue, ValueDecoderError{
			Name:     "UintDecodeValue",
			Kinds:    []reflect.Kind{reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint},
			Received: reflect.Zero(t),
		}
	}
}

// DecodeValue is the ValueDecoder for uint types.
func (uic *UIntCodec) DecodeValue(dc DecodeContext, vr ValueReader, val reflect.Value) error {
	if !val.CanSet() {
		return ValueDecoderError{
			Name:     "UintDecodeValue",
			Kinds:    []reflect.Kind{reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint},
			Received: val,
		}
	}

	elem, err := uic.decodeType(dc, vr, val.Type())
	if err != nil {
		return err
	}

	val.SetUint(elem.Uint())
	return nil
}
