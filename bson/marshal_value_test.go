// Copyright (C) MongoDB, Inc. 2017-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package bson

import (
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/internal/assert"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func TestMarshalValue(t *testing.T) {
	t.Parallel()

	marshalValueTestCases := newMarshalValueTestCasesWithInterfaceCore(t)

	t.Run("MarshalValue", func(t *testing.T) {
		t.Parallel()

		for _, tc := range marshalValueTestCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				valueType, valueBytes, err := MarshalValue(tc.val)
				assert.Nil(t, err, "MarshalValue error: %v", err)
				compareMarshalValueResults(t, tc, valueType, valueBytes)
			})
		}
	})
	t.Run("MarshalValueAppend", func(t *testing.T) {
		t.Parallel()

		for _, tc := range marshalValueTestCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				valueType, valueBytes, err := MarshalValueAppend(nil, tc.val)
				assert.Nil(t, err, "MarshalValueAppend error: %v", err)
				compareMarshalValueResults(t, tc, valueType, valueBytes)
			})
		}
	})
	t.Run("MarshalValueWithRegistry", func(t *testing.T) {
		t.Parallel()

		for _, tc := range marshalValueTestCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				valueType, valueBytes, err := MarshalValueWithRegistry(DefaultRegistry, tc.val)
				assert.Nil(t, err, "MarshalValueWithRegistry error: %v", err)
				compareMarshalValueResults(t, tc, valueType, valueBytes)
			})
		}
	})
	t.Run("MarshalValueWithContext", func(t *testing.T) {
		t.Parallel()

		ec := bsoncodec.EncodeContext{Registry: DefaultRegistry}
		for _, tc := range marshalValueTestCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				valueType, valueBytes, err := MarshalValueWithContext(ec, tc.val)
				assert.Nil(t, err, "MarshalValueWithContext error: %v", err)
				compareMarshalValueResults(t, tc, valueType, valueBytes)
			})
		}
	})
	t.Run("MarshalValueAppendWithRegistry", func(t *testing.T) {
		t.Parallel()

		for _, tc := range marshalValueTestCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				valueType, valueBytes, err := MarshalValueAppendWithRegistry(DefaultRegistry, nil, tc.val)
				assert.Nil(t, err, "MarshalValueAppendWithRegistry error: %v", err)
				compareMarshalValueResults(t, tc, valueType, valueBytes)
			})
		}
	})
	t.Run("MarshalValueAppendWithContext", func(t *testing.T) {
		t.Parallel()

		ec := bsoncodec.EncodeContext{Registry: DefaultRegistry}
		for _, tc := range marshalValueTestCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				valueType, valueBytes, err := MarshalValueAppendWithContext(ec, nil, tc.val)
				assert.Nil(t, err, "MarshalValueWithContext error: %v", err)
				compareMarshalValueResults(t, tc, valueType, valueBytes)
			})
		}
	})
}

func BenchmarkMarshalValueAppendWithContext(b *testing.B) {
	var (
		oid           = primitive.NewObjectID()
		regex         = primitive.Regex{Pattern: "pattern", Options: "imx"}
		dbPointer     = primitive.DBPointer{DB: "db", Pointer: primitive.NewObjectID()}
		codeWithScope = primitive.CodeWithScope{Code: "code", Scope: D{{"a", "b"}}}
		decimal128    = primitive.NewDecimal128(5, 10)
		structTest    = marshalValueStruct{Foo: 10}
	)
	idx, scopeCore := bsoncore.AppendDocumentStart(nil)
	scopeCore = bsoncore.AppendStringElement(scopeCore, "a", "b")
	scopeCore, _ = bsoncore.AppendDocumentEnd(scopeCore, idx)
	structCore, _ := Marshal(structTest)

	marshalValueTestCases := []marshalValueTestCase{
		{"double", 3.14, bsontype.Double, bsoncore.AppendDouble(nil, 3.14)},
		{"string", "hello world", bsontype.String, bsoncore.AppendString(nil, "hello world")},
		{"binary", primitive.Binary{1, []byte{1, 2}}, bsontype.Binary, bsoncore.AppendBinary(nil, 1, []byte{1, 2})},
		{"undefined", primitive.Undefined{}, bsontype.Undefined, []byte{}},
		{"object id", oid, bsontype.ObjectID, bsoncore.AppendObjectID(nil, oid)},
		{"boolean", true, bsontype.Boolean, bsoncore.AppendBoolean(nil, true)},
		{"datetime", primitive.DateTime(5), bsontype.DateTime, bsoncore.AppendDateTime(nil, 5)},
		{"null", primitive.Null{}, bsontype.Null, []byte{}},
		{"regex", regex, bsontype.Regex, bsoncore.AppendRegex(nil, regex.Pattern, regex.Options)},
		{"dbpointer", dbPointer, bsontype.DBPointer, bsoncore.AppendDBPointer(nil, dbPointer.DB, dbPointer.Pointer)},
		{"javascript", primitive.JavaScript("js"), bsontype.JavaScript, bsoncore.AppendJavaScript(nil, "js")},
		{"symbol", primitive.Symbol("symbol"), bsontype.Symbol, bsoncore.AppendSymbol(nil, "symbol")},
		{"code with scope", codeWithScope, bsontype.CodeWithScope, bsoncore.AppendCodeWithScope(nil, "code", scopeCore)},
		{"int32", 5, bsontype.Int32, bsoncore.AppendInt32(nil, 5)},
		{"int64", int64(5), bsontype.Int64, bsoncore.AppendInt64(nil, 5)},
		{"timestamp", primitive.Timestamp{T: 1, I: 5}, bsontype.Timestamp, bsoncore.AppendTimestamp(nil, 1, 5)},
		{"decimal128", decimal128, bsontype.Decimal128, bsoncore.AppendDecimal128(nil, decimal128)},
		{"min key", primitive.MinKey{}, bsontype.MinKey, []byte{}},
		{"max key", primitive.MaxKey{}, bsontype.MaxKey, []byte{}},
		{"struct", structTest, bsontype.EmbeddedDocument, structCore},
		{"D", D{{"foo", int32(10)}}, bsontype.EmbeddedDocument, structCore},
		{"M", M{"foo": int32(10)}, bsontype.EmbeddedDocument, structCore},
		{"ValueMarshaler", marshalValueMarshaler{Foo: 10}, bsontype.Int32, bsoncore.AppendInt32(nil, 10)},
	}

	ec := bsoncodec.EncodeContext{Registry: DefaultRegistry}
	for n := 0; n < b.N; n++ {
		for _, tc := range marshalValueTestCases {
			_, _, err := MarshalValueAppendWithContext(ec, nil, tc.val)
			if err != nil {
				b.Fatal()
			}
		}
	}
}

func compareMarshalValueResults(t *testing.T, tc marshalValueTestCase, gotType bsontype.Type, gotBytes []byte) {
	t.Helper()
	expectedValue := RawValue{Type: tc.bsontype, Value: tc.bytes}
	gotValue := RawValue{Type: gotType, Value: gotBytes}
	assert.Equal(t, expectedValue, gotValue, "value mismatch; expected %s, got %s", expectedValue, gotValue)
}

// benchmark covering GODRIVER-2779
func BenchmarkSliceCodecMarshal(b *testing.B) {
	testStruct := unmarshalerNonPtrStruct{B: []byte(strings.Repeat("t", 4096))}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _, err := MarshalValue(testStruct)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
