// Copyright (C) MongoDB, Inc. 2017-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package connstring_test

import (
	"fmt"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/internal/assert"
	"go.mongodb.org/mongo-driver/v2/internal/require"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/connstring"
)

func TestAppName(t *testing.T) {
	tests := []struct {
		s        string
		expected string
		err      bool
	}{
		{s: "appName=Funny", expected: "Funny"},
		{s: "appName=awesome", expected: "awesome"},
		{s: "appName=", expected: ""},
	}

	for _, test := range tests {
		s := fmt.Sprintf("mongodb://localhost/?%s", test.s)
		t.Run(s, func(t *testing.T) {
			cs, err := connstring.ParseAndValidate(s)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, cs.AppName)
			}
		})
	}
}

func TestAuthMechanism(t *testing.T) {
	tests := []struct {
		s        string
		expected string
		err      bool
	}{
		{s: "authMechanism=scram-sha-1", expected: "scram-sha-1"},
		{s: "authMechanism=scram-sha-256", expected: "scram-sha-256"},
		{s: "authMechanism=plain", expected: "plain"},
	}

	for _, test := range tests {
		s := fmt.Sprintf("mongodb://user:pass@localhost/?%s", test.s)
		t.Run(s, func(t *testing.T) {
			cs, err := connstring.ParseAndValidate(s)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, cs.AuthMechanism)
			}
		})
	}
}

func TestAuthSource(t *testing.T) {
	tests := []struct {
		s        string
		expected string
		err      bool
	}{
		{s: "foobar?authSource=bazqux", expected: "bazqux"},
		{s: "foobar", expected: "foobar"},
		{s: "", expected: "admin"},
	}

	for _, test := range tests {
		s := fmt.Sprintf("mongodb://user:pass@localhost/%s", test.s)
		t.Run(s, func(t *testing.T) {
			cs, err := connstring.ParseAndValidate(s)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, cs.AuthSource)
			}
		})
	}

	tests = []struct {
		s        string
		expected string
		err      bool
	}{
		{s: "authMechanismProperties=ENVIRONMENT:gcp,TOKEN_RESOURCE:mongodb://test-cluster", expected: "$external"},
	}

	for _, test := range tests {
		s := fmt.Sprintf("mongodb://test.mongodb.net/?authMechanism=MONGODB-OIDC&/%s", test.s)
		t.Run(s, func(t *testing.T) {
			cs, err := connstring.ParseAndValidate(s)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, cs.AuthSource)
			}
		})
	}

}

func TestConnect(t *testing.T) {
	tests := []struct {
		s        string
		expected connstring.ConnectMode
		err      bool
	}{
		{s: "connect=automatic", expected: connstring.AutoConnect},
		{s: "connect=AUTOMATIC", expected: connstring.AutoConnect},
		{s: "connect=direct", expected: connstring.SingleConnect},
		{s: "connect=blah", err: true},
		// Combinations of connect and directConnection where connect is set first - conflicting combinations must
		// error.
		{s: "connect=automatic&directConnection=true", err: true},
		{s: "connect=automatic&directConnection=false", expected: connstring.AutoConnect},
		{s: "connect=direct&directConnection=true", expected: connstring.SingleConnect},
		{s: "connect=direct&directConnection=false", err: true},
		// Combinations of connect and directConnection where directConnection is set first.
		{s: "directConnection=true&connect=automatic", err: true},
		{s: "directConnection=false&connect=automatic", expected: connstring.AutoConnect},
		{s: "directConnection=true&connect=direct", expected: connstring.SingleConnect},
		{s: "directConnection=false&connect=direct", err: true},
	}

	for _, test := range tests {
		s := fmt.Sprintf("mongodb://localhost/?%s", test.s)
		t.Run(s, func(t *testing.T) {
			cs, err := connstring.ParseAndValidate(s)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, cs.Connect)
			}
		})
	}
}

func TestDirectConnection(t *testing.T) {
	testCases := []struct {
		s        string
		expected bool
		err      bool
	}{
		{"directConnection=true", true, false},
		{"directConnection=false", false, false},
		{"directConnection=TRUE", true, false},
		{"directConnection=FALSE", false, false},
		{"directConnection=blah", false, true},
	}

	for _, tc := range testCases {
		s := fmt.Sprintf("mongodb://localhost/?%s", tc.s)
		t.Run(s, func(t *testing.T) {
			cs, err := connstring.ParseAndValidate(s)
			if tc.err {
				assert.NotNil(t, err, "expected error, got nil")
				return
			}

			assert.Nil(t, err, "expected no error, got %v", err)
			assert.Equal(t, tc.expected, cs.DirectConnection, "expected DirectConnection value %v, got %v", tc.expected,
				cs.DirectConnection)
			assert.True(t, cs.DirectConnectionSet, "expected DirectConnectionSet to be true, got false")
		})
	}
}

func TestConnectTimeout(t *testing.T) {
	tests := []struct {
		s        string
		expected time.Duration
		err      bool
	}{
		{s: "connectTimeoutMS=10", expected: time.Duration(10) * time.Millisecond},
		{s: "connectTimeoutMS=100", expected: time.Duration(100) * time.Millisecond},
		{s: "connectTimeoutMS=-2", err: true},
		{s: "connectTimeoutMS=gsdge", err: true},
	}

	for _, test := range tests {
		s := fmt.Sprintf("mongodb://localhost/?%s", test.s)
		t.Run(s, func(t *testing.T) {
			cs, err := connstring.ParseAndValidate(s)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, cs.ConnectTimeout)
				require.True(t, cs.ConnectTimeoutSet)
			}
		})
	}
}

func TestHeartbeatInterval(t *testing.T) {
	tests := []struct {
		s        string
		expected time.Duration
		err      bool
	}{
		{s: "heartbeatIntervalMS=10", expected: time.Duration(10) * time.Millisecond},
		{s: "heartbeatIntervalMS=100", expected: time.Duration(100) * time.Millisecond},
		{s: "heartbeatIntervalMS=-2", err: true},
		{s: "heartbeatIntervalMS=gsdge", err: true},
	}

	for _, test := range tests {
		s := fmt.Sprintf("mongodb://localhost/?%s", test.s)
		t.Run(s, func(t *testing.T) {
			cs, err := connstring.ParseAndValidate(s)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, cs.HeartbeatInterval)
			}
		})
	}
}

func TestLocalThreshold(t *testing.T) {
	tests := []struct {
		s        string
		expected time.Duration
		err      bool
	}{
		{s: "localThresholdMS=0", expected: time.Duration(0) * time.Millisecond},
		{s: "localThresholdMS=10", expected: time.Duration(10) * time.Millisecond},
		{s: "localThresholdMS=100", expected: time.Duration(100) * time.Millisecond},
		{s: "localThresholdMS=-2", err: true},
		{s: "localThresholdMS=gsdge", err: true},
	}

	for _, test := range tests {
		s := fmt.Sprintf("mongodb://localhost/?%s", test.s)
		t.Run(s, func(t *testing.T) {
			cs, err := connstring.ParseAndValidate(s)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, cs.LocalThreshold)
			}
		})
	}
}

func TestMaxConnIdleTime(t *testing.T) {
	tests := []struct {
		s        string
		expected time.Duration
		err      bool
	}{
		{s: "maxIdleTimeMS=10", expected: time.Duration(10) * time.Millisecond},
		{s: "maxIdleTimeMS=100", expected: time.Duration(100) * time.Millisecond},
		{s: "maxIdleTimeMS=-2", err: true},
		{s: "maxIdleTimeMS=gsdge", err: true},
	}

	for _, test := range tests {
		s := fmt.Sprintf("mongodb://localhost/?%s", test.s)
		t.Run(s, func(t *testing.T) {
			cs, err := connstring.ParseAndValidate(s)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, cs.MaxConnIdleTime)
			}
		})
	}
}

func TestMaxPoolSize(t *testing.T) {
	tests := []struct {
		s        string
		expected uint64
		err      bool
	}{
		{s: "maxPoolSize=10", expected: 10},
		{s: "maxPoolSize=100", expected: 100},
		{s: "maxPoolSize=-2", err: true},
		{s: "maxPoolSize=gsdge", err: true},
	}

	for _, test := range tests {
		s := fmt.Sprintf("mongodb://localhost/?%s", test.s)
		t.Run(s, func(t *testing.T) {
			cs, err := connstring.ParseAndValidate(s)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.True(t, cs.MaxPoolSizeSet)
				require.Equal(t, test.expected, cs.MaxPoolSize)
			}
		})
	}
}

func TestMinPoolSize(t *testing.T) {
	tests := []struct {
		s        string
		expected uint64
		err      bool
	}{
		{s: "minPoolSize=10", expected: 10},
		{s: "minPoolSize=100", expected: 100},
		{s: "minPoolSize=-2", err: true},
		{s: "minPoolSize=gsdge", err: true},
	}

	for _, test := range tests {
		s := fmt.Sprintf("mongodb://localhost/?%s", test.s)
		t.Run(s, func(t *testing.T) {
			cs, err := connstring.ParseAndValidate(s)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.True(t, cs.MinPoolSizeSet)
				require.Equal(t, test.expected, cs.MinPoolSize)
			}
		})
	}
}

func TestMaxConnecting(t *testing.T) {
	tests := []struct {
		s        string
		expected uint64
		err      bool
	}{
		{s: "maxConnecting=10", expected: 10},
		{s: "maxConnecting=100", expected: 100},
		{s: "maxConnecting=-2", err: true},
		{s: "maxConnecting=gsdge", err: true},
	}

	for _, test := range tests {
		s := fmt.Sprintf("mongodb://localhost/?%s", test.s)
		t.Run(s, func(t *testing.T) {
			cs, err := connstring.ParseAndValidate(s)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.True(t, cs.MaxConnectingSet)
				require.Equal(t, test.expected, cs.MaxConnecting)
			}
		})
	}
}

func TestReadPreference(t *testing.T) {
	tests := []struct {
		s        string
		expected string
		err      bool
	}{
		{s: "readPreference=primary", expected: "primary"},
		{s: "readPreference=secondaryPreferred", expected: "secondaryPreferred"},
		{s: "readPreference=something", expected: "something"}, // we don't validate here
	}

	for _, test := range tests {
		s := fmt.Sprintf("mongodb://localhost/?%s", test.s)
		t.Run(s, func(t *testing.T) {
			cs, err := connstring.ParseAndValidate(s)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, cs.ReadPreference)
			}
		})
	}
}

func TestReadPreferenceTags(t *testing.T) {
	tests := []struct {
		s        string
		expected []map[string]string
		err      bool
	}{
		{s: "", expected: nil},
		{s: "readPreferenceTags=one:1", expected: []map[string]string{{"one": "1"}}},
		{s: "readPreferenceTags=one:1,two:2", expected: []map[string]string{{"one": "1", "two": "2"}}},
		{s: "readPreferenceTags=one:1&readPreferenceTags=two:2", expected: []map[string]string{{"one": "1"}, {"two": "2"}}},
		{s: "readPreferenceTags=one:1:3,two:2", err: true},
		{s: "readPreferenceTags=one:1&readPreferenceTags=two:2&readPreferenceTags=", expected: []map[string]string{{"one": "1"}, {"two": "2"}, {}}},
		{s: "readPreferenceTags=", expected: []map[string]string{{}}},
	}

	for _, test := range tests {
		s := fmt.Sprintf("mongodb://localhost/?%s", test.s)
		t.Run(s, func(t *testing.T) {
			cs, err := connstring.ParseAndValidate(s)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, cs.ReadPreferenceTagSets)
			}
		})
	}
}

func TestMaxStaleness(t *testing.T) {
	tests := []struct {
		s        string
		expected time.Duration
		err      bool
	}{
		{s: "maxStaleness=10", expected: time.Duration(10) * time.Second},
		{s: "maxStaleness=100", expected: time.Duration(100) * time.Second},
		{s: "maxStaleness=-2", err: true},
		{s: "maxStaleness=gsdge", err: true},
	}
	for _, test := range tests {
		s := fmt.Sprintf("mongodb://localhost/?%s", test.s)
		t.Run(s, func(t *testing.T) {
			cs, err := connstring.ParseAndValidate(s)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, cs.MaxStaleness)
			}
		})
	}
}

func TestReplicaSet(t *testing.T) {
	tests := []struct {
		s        string
		expected string
		err      bool
	}{
		{s: "replicaSet=auto", expected: "auto"},
		{s: "replicaSet=rs0", expected: "rs0"},
	}

	for _, test := range tests {
		s := fmt.Sprintf("mongodb://localhost/?%s", test.s)
		t.Run(s, func(t *testing.T) {
			cs, err := connstring.ParseAndValidate(s)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, cs.ReplicaSet)
			}
		})
	}
}

func TestRetryWrites(t *testing.T) {
	tests := []struct {
		s        string
		expected bool
		err      bool
	}{
		{s: "retryWrites=true", expected: true},
		{s: "retryWrites=false", expected: false},
		{s: "retryWrites=foobar", expected: false, err: true},
	}

	for _, test := range tests {
		s := fmt.Sprintf("mongodb://localhost/?%s", test.s)
		t.Run(s, func(t *testing.T) {
			cs, err := connstring.ParseAndValidate(s)
			if test.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.expected, cs.RetryWrites)
			require.True(t, cs.RetryWritesSet)
		})
	}
}

func TestRetryReads(t *testing.T) {
	tests := []struct {
		s        string
		expected bool
		err      bool
	}{
		{s: "retryReads=true", expected: true},
		{s: "retryReads=false", expected: false},
		{s: "retryReads=foobar", expected: false, err: true},
	}

	for _, test := range tests {
		s := fmt.Sprintf("mongodb://localhost/?%s", test.s)
		t.Run(s, func(t *testing.T) {
			cs, err := connstring.ParseAndValidate(s)
			if test.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.expected, cs.RetryReads)
			require.True(t, cs.RetryReadsSet)
		})
	}
}

func TestScheme(t *testing.T) {
	// Can't unit test 'mongodb+srv' because that requires networking.  Tested
	// in x/mongo/driver/topology/initial_dns_seedlist_discovery_test.go
	cs, err := connstring.ParseAndValidate("mongodb://localhost/")
	require.NoError(t, err)
	require.Equal(t, cs.Scheme, "mongodb")
	require.Equal(t, cs.Scheme, connstring.SchemeMongoDB)
}

func TestServerSelectionTimeout(t *testing.T) {
	tests := []struct {
		s        string
		expected time.Duration
		err      bool
	}{
		{s: "serverSelectionTimeoutMS=10", expected: time.Duration(10) * time.Millisecond},
		{s: "serverSelectionTimeoutMS=100", expected: time.Duration(100) * time.Millisecond},
		{s: "serverSelectionTimeoutMS=-2", err: true},
		{s: "serverSelectionTimeoutMS=gsdge", err: true},
	}

	for _, test := range tests {
		s := fmt.Sprintf("mongodb://localhost/?%s", test.s)
		t.Run(s, func(t *testing.T) {
			cs, err := connstring.ParseAndValidate(s)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, cs.ServerSelectionTimeout)
				require.True(t, cs.ServerSelectionTimeoutSet)
			}
		})
	}
}

func TestSocketTimeout(t *testing.T) {
	tests := []struct {
		s        string
		expected time.Duration
		err      bool
	}{
		{s: "socketTimeoutMS=10", expected: time.Duration(10) * time.Millisecond},
		{s: "socketTimeoutMS=100", expected: time.Duration(100) * time.Millisecond},
		{s: "socketTimeoutMS=-2", err: true},
		{s: "socketTimeoutMS=gsdge", err: true},
	}

	for _, test := range tests {
		s := fmt.Sprintf("mongodb://localhost/?%s", test.s)
		t.Run(s, func(t *testing.T) {
			cs, err := connstring.ParseAndValidate(s)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, cs.SocketTimeout)
				require.True(t, cs.SocketTimeoutSet)
			}
		})
	}
}

func TestCompressionOptions(t *testing.T) {
	tests := []struct {
		name        string
		uriOptions  string
		compressors []string
		zlibLevel   int
		zstdLevel   int
		err         bool
	}{
		{name: "SingleCompressor", uriOptions: "compressors=zlib", compressors: []string{"zlib"}},
		{name: "MultiCompressors", uriOptions: "compressors=snappy,zlib", compressors: []string{"snappy", "zlib"}},
		{name: "ZlibWithLevel", uriOptions: "compressors=zlib&zlibCompressionLevel=7", compressors: []string{"zlib"}, zlibLevel: 7},
		{name: "DefaultZlibLevel", uriOptions: "compressors=zlib&zlibCompressionLevel=-1", compressors: []string{"zlib"}, zlibLevel: 6},
		{name: "InvalidZlibLevel", uriOptions: "compressors=zlib&zlibCompressionLevel=-2", compressors: []string{"zlib"}, err: true},
		{name: "ZstdWithLevel", uriOptions: "compressors=zstd&zstdCompressionLevel=20", compressors: []string{"zstd"}, zstdLevel: 20},
		{name: "DefaultZstdLevel", uriOptions: "compressors=zstd&zstdCompressionLevel=-1", compressors: []string{"zstd"}, zstdLevel: 6},
		{name: "InvalidZstdLevel", uriOptions: "compressors=zstd&zstdCompressionLevel=30", compressors: []string{"zstd"}, err: true},
	}

	for _, tc := range tests {
		uri := fmt.Sprintf("mongodb://localhost/?%s", tc.uriOptions)
		t.Run(tc.name, func(t *testing.T) {
			cs, err := connstring.ParseAndValidate(uri)
			if tc.err {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.compressors, cs.Compressors)
				if tc.zlibLevel != 0 {
					assert.Equal(t, tc.zlibLevel, cs.ZlibLevel)
				}
				if tc.zstdLevel != 0 {
					assert.Equal(t, tc.zstdLevel, cs.ZstdLevel)
				}
			}
		})
	}
}
