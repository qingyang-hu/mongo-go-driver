// Copyright (C) MongoDB, Inc. 2017-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/internal/aws/credentials"
	v4signer "go.mongodb.org/mongo-driver/v2/internal/aws/signer/v4"
	"go.mongodb.org/mongo-driver/v2/internal/credproviders"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/auth/creds"
)

// MongoDBAWS is the mechanism name for MongoDBAWS.
const MongoDBAWS = "MONGODB-AWS"

func newMongoDBAWSAuthenticator(cred *Cred, httpClient *http.Client) (Authenticator, error) {
	if cred.Source != "" && cred.Source != sourceExternal {
		return nil, newAuthError("MONGODB-AWS source must be empty or $external", nil)
	}

	if cred.AWSSigner != nil {
		return &MongoDBAWSAuthenticator{
			signer: cred.AWSSigner,
		}, nil
	}

	if httpClient == nil {
		return nil, errors.New("httpClient must not be nil when AWSSigner is not provided in cred")
	}
	credentials := &credproviders.StaticProvider{
		Value: credentials.Value{
			AccessKeyID:     cred.Username,
			SecretAccessKey: cred.Password,
			SessionToken:    cred.Props["AWS_SESSION_TOKEN"],
		},
	}
	providers := creds.NewAWSCredentialProvider(httpClient, credentials)
	return &MongoDBAWSAuthenticator{
		signer: &builtInV4Signer{
			credentials: providers.Cred,
		},
	}, nil
}

// MongoDBAWSAuthenticator uses AWS-IAM credentials over SASL to authenticate a connection.
type MongoDBAWSAuthenticator struct {
	signer driver.AWSSigner
}

// Auth authenticates the connection.
func (a *MongoDBAWSAuthenticator) Auth(ctx context.Context, cfg *driver.AuthConfig) error {
	awsSasl := &awsSaslAdapter{
		signer: a.signer,
	}
	err := ConductSaslConversation(ctx, cfg, sourceExternal, awsSasl)
	if err != nil {
		return newAuthError("sasl conversation error", err)
	}
	return nil
}

// Reauth reauthenticates the connection.
func (a *MongoDBAWSAuthenticator) Reauth(_ context.Context, _ *driver.AuthConfig) error {
	return newAuthError("AWS authentication does not support reauthentication", nil)
}

var _ driver.AWSSigner = (*builtInV4Signer)(nil)

type builtInV4Signer struct {
	credentials *credentials.Credentials
}

func (b *builtInV4Signer) Sign(ctx context.Context, newReq func(string) *http.Request, body, service, region string, signTime time.Time) error {
	creds, err := b.credentials.GetWithContext(ctx)
	if err != nil {
		return err
	}
	req := newReq(creds.SessionToken)
	signer := v4signer.NewSigner(b.credentials)
	_, err = signer.Sign(req, strings.NewReader(body), service, region, signTime)
	return err
}
