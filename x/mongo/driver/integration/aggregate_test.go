// Copyright (C) MongoDB, Inc. 2022-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package integration

import (
	"bytes"
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/internal/integtest"
	"go.mongodb.org/mongo-driver/v2/internal/require"
	"go.mongodb.org/mongo-driver/v2/internal/serverselector"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
	"go.mongodb.org/mongo-driver/v2/x/bsonx/bsoncore"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/operation"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/topology"
)

func setUpMonitor() (*event.CommandMonitor, chan *event.CommandStartedEvent, chan *event.CommandSucceededEvent, chan *event.CommandFailedEvent) {
	started := make(chan *event.CommandStartedEvent, 1)
	succeeded := make(chan *event.CommandSucceededEvent, 1)
	failed := make(chan *event.CommandFailedEvent, 1)

	return &event.CommandMonitor{
		Started: func(ctx context.Context, e *event.CommandStartedEvent) {
			started <- e
		},
		Succeeded: func(ctx context.Context, e *event.CommandSucceededEvent) {
			succeeded <- e
		},
		Failed: func(ctx context.Context, e *event.CommandFailedEvent) {
			failed <- e
		},
	}, started, succeeded, failed
}

func skipIfBelow32(ctx context.Context, t *testing.T, topo *topology.Topology) {
	server, err := topo.SelectServer(ctx, &serverselector.Write{})
	noerr(t, err)

	versionCmd := bsoncore.BuildDocument(nil, bsoncore.AppendInt32Element(nil, "serverStatus", 1))
	serverStatus, err := runCommand(server, dbName, versionCmd)
	noerr(t, err)
	version, err := serverStatus.LookupErr("version")
	noerr(t, err)

	if integtest.CompareVersions(t, version.StringValue(), "3.2") < 0 {
		t.Skip()
	}
}

func TestAggregate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("TestMaxTimeMSInGetMore", func(t *testing.T) {
		ctx := context.Background()
		monitor, started, succeeded, failed := setUpMonitor()
		dbName := "TestAggMaxTimeDB"
		collName := "TestAggMaxTimeColl"
		top := integtest.MonitoredTopology(t, dbName, monitor)
		clearChannels(started, succeeded, failed)
		skipIfBelow32(ctx, t, top)

		clearChannels(started, succeeded, failed)
		err := operation.NewInsert(
			bsoncore.BuildDocument(nil, bsoncore.AppendInt32Element(nil, "x", 1)),
			bsoncore.BuildDocument(nil, bsoncore.AppendInt32Element(nil, "x", 1)),
			bsoncore.BuildDocument(nil, bsoncore.AppendInt32Element(nil, "x", 1)),
		).Collection(collName).Database(dbName).
			Deployment(top).ServerSelector(&serverselector.Write{}).Execute(context.Background())
		noerr(t, err)

		clearChannels(started, succeeded, failed)
		op := operation.NewAggregate(bsoncore.BuildDocumentFromElements(nil)).
			Collection(collName).Database(dbName).Deployment(top).ServerSelector(&serverselector.Write{}).
			CommandMonitor(monitor).BatchSize(2)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err = op.Execute(ctx)
		noerr(t, err)
		batchCursor, err := op.Result(driver.CursorOptions{BatchSize: 2, CommandMonitor: monitor})
		noerr(t, err)

		var e *event.CommandStartedEvent
		select {
		case e = <-started:
		case <-time.After(2000 * time.Millisecond):
			t.Fatal("timed out waiting for aggregate")
		}

		require.Equal(t, "aggregate", e.CommandName)

		clearChannels(started, succeeded, failed)
		// first Next() should automatically return true
		require.True(t, batchCursor.Next(ctx), "expected true from first Next, got false")
		clearChannels(started, succeeded, failed)
		batchCursor.Next(ctx) // should do getMore

		select {
		case e = <-started:
		case <-time.After(200 * time.Millisecond):
			t.Fatal("timed out waiting for getMore")
		}
		require.Equal(t, "getMore", e.CommandName)
		_, err = e.Command.LookupErr("maxTimeMS")
		noerr(t, err)
	})
	t.Run("Multiple Batches", func(t *testing.T) {
		ds := []bsoncore.Document{
			bsoncore.BuildDocument(nil, bsoncore.AppendInt32Element(nil, "_id", 1)),
			bsoncore.BuildDocument(nil, bsoncore.AppendInt32Element(nil, "_id", 2)),
			bsoncore.BuildDocument(nil, bsoncore.AppendInt32Element(nil, "_id", 3)),
			bsoncore.BuildDocument(nil, bsoncore.AppendInt32Element(nil, "_id", 4)),
			bsoncore.BuildDocument(nil, bsoncore.AppendInt32Element(nil, "_id", 5)),
		}
		wc := writeconcern.Majority()
		autoInsertDocs(t, wc, ds...)

		op := operation.NewAggregate(bsoncore.BuildArray(nil,
			bsoncore.BuildDocumentValue(
				bsoncore.BuildDocumentElement(nil,
					"$match", bsoncore.BuildDocumentElement(nil,
						"_id", bsoncore.AppendInt32Element(nil, "$gt", 2),
					),
				),
			),
			bsoncore.BuildDocumentValue(
				bsoncore.BuildDocumentElement(nil,
					"$sort", bsoncore.AppendInt32Element(nil, "_id", 1),
				),
			),
		)).Collection(integtest.ColName(t)).Database(dbName).Deployment(integtest.Topology(t)).
			ServerSelector(&serverselector.Write{}).BatchSize(2)
		err := op.Execute(context.Background())
		noerr(t, err)
		cursor, err := op.Result(driver.CursorOptions{BatchSize: 2})
		noerr(t, err)

		var got []bsoncore.Document
		for i := 0; i < 2; i++ {
			if !cursor.Next(context.Background()) {
				t.Error("Cursor should have results, but does not have a next result")
			}
			docs, err := cursor.Batch().Documents()
			noerr(t, err)
			got = append(got, docs...)
		}
		readers := ds[2:]
		for i, g := range got {
			if !bytes.Equal(g[:len(readers[i])], readers[i]) {
				t.Errorf("Did not get expected document. got %v; want %v", bson.Raw(g[:len(readers[i])]), readers[i])
			}
		}

		if cursor.Next(context.Background()) {
			t.Error("Cursor should be exhausted but has more results")
		}
	})
	t.Run("AllowDiskUse", func(t *testing.T) {
		ds := []bsoncore.Document{
			bsoncore.BuildDocument(nil, bsoncore.AppendInt32Element(nil, "_id", 1)),
			bsoncore.BuildDocument(nil, bsoncore.AppendInt32Element(nil, "_id", 2)),
		}
		wc := writeconcern.Majority()
		autoInsertDocs(t, wc, ds...)

		op := operation.NewAggregate(bsoncore.BuildArray(nil)).Collection(integtest.ColName(t)).Database(dbName).
			Deployment(integtest.Topology(t)).ServerSelector(&serverselector.Write{}).AllowDiskUse(true)
		err := op.Execute(context.Background())
		if err != nil {
			t.Errorf("Expected no error from allowing disk use, but got %v", err)
		}
	})

}

func clearChannels(s chan *event.CommandStartedEvent, succ chan *event.CommandSucceededEvent, f chan *event.CommandFailedEvent) {
	for len(s) > 0 {
		<-s
	}
	for len(succ) > 0 {
		<-succ
	}
	for len(f) > 0 {
		<-f
	}
}
