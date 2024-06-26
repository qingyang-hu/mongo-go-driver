// Copyright (C) MongoDB, Inc. 2017-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package options

import (
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// TransactionOptions represents options that can be used to configure a transaction.
type TransactionOptions struct {
	// The read concern for operations in the transaction. The default value is nil, which means that the default
	// read concern of the session used to start the transaction will be used.
	ReadConcern *readconcern.ReadConcern

	// The read preference for operations in the transaction. The default value is nil, which means that the default
	// read preference of the session used to start the transaction will be used.
	ReadPreference *readpref.ReadPref

	// The write concern for operations in the transaction. The default value is nil, which means that the default
	// write concern of the session used to start the transaction will be used.
	WriteConcern *writeconcern.WriteConcern
}

// Transaction creates a new TransactionOptions instance.
func Transaction() *TransactionOptions {
	return &TransactionOptions{}
}

// SetReadConcern sets the value for the ReadConcern field.
func (t *TransactionOptions) SetReadConcern(rc *readconcern.ReadConcern) *TransactionOptions {
	t.ReadConcern = rc
	return t
}

// SetReadPreference sets the value for the ReadPreference field.
func (t *TransactionOptions) SetReadPreference(rp *readpref.ReadPref) *TransactionOptions {
	t.ReadPreference = rp
	return t
}

// SetWriteConcern sets the value for the WriteConcern field.
func (t *TransactionOptions) SetWriteConcern(wc *writeconcern.WriteConcern) *TransactionOptions {
	t.WriteConcern = wc
	return t
}
