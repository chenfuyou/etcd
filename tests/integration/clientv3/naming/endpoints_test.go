// Copyright 2016 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package naming_test

import (
	"context"
	"reflect"
	"testing"

	etcd "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"

	"go.etcd.io/etcd/pkg/v3/testutil"
	"go.etcd.io/etcd/tests/v3/integration"
)

func TestEndpointManager(t *testing.T) {
	t.Skip("Not implemented yet")

	defer testutil.AfterTest(t)

	clus := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	em, err := endpoints.NewManager(clus.RandClient(), "foo")
	if err != nil {
		t.Fatal("failed to create EndpointManager", err)
	}
	ctx, watchCancel := context.WithCancel(context.Background())
	defer watchCancel()
	w, err := em.NewWatchChannel(ctx)
	if err != nil {
		t.Fatal("failed to establish watch", err)
	}

	e1 := endpoints.Endpoint{Addr: "127.0.0.1", Metadata: "metadata"}
	err = em.AddEndpoint(context.TODO(), "foo/a1", e1)
	if err != nil {
		t.Fatal("failed to add foo", err)
	}

	us := <-w

	if us == nil {
		t.Fatal("failed to get update", err)
	}

	wu := endpoints.Update{
		Op:       endpoints.Add,
		Key:      "foo/a1",
		Endpoint: e1,
	}

	if !reflect.DeepEqual(us[0], wu) {
		t.Fatalf("up = %#v, want %#v", us[0], wu)
	}

	err = em.DeleteEndpoint(context.TODO(), "foo/a1")
	if err != nil {
		t.Fatalf("failed to udpate %v", err)
	}

	us = <-w
	if err != nil {
		t.Fatalf("failed to get udpate %v", err)
	}

	wu = endpoints.Update{
		Op:  endpoints.Delete,
		Key: "foo/a1",
	}

	if !reflect.DeepEqual(us, wu) {
		t.Fatalf("up = %#v, want %#v", us[1], wu)
	}
}

// TestEndpointManagerAtomicity ensures the resolver will initialize
// correctly with multiple hosts and correctly receive multiple
// updates in a single revision.
func TestEndpointManagerAtomicity(t *testing.T) {
	t.Skip("Not implemented yet")

	defer testutil.AfterTest(t)

	clus := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	c := clus.RandClient()
	em, err := endpoints.NewManager(c, "foo")
	if err != nil {
		t.Fatal("failed to create EndpointManager", err)
	}

	err = em.Update(context.TODO(), []*endpoints.UpdateWithOpts{
		endpoints.NewAddUpdateOpts("foo/host", endpoints.Endpoint{Addr: "127.0.0.1:2000"}),
		endpoints.NewAddUpdateOpts("foo/host2", endpoints.Endpoint{Addr: "127.0.0.1:2001"})})
	if err != nil {
		t.Fatal(err)
	}

	ctx, watchCancel := context.WithCancel(context.Background())
	defer watchCancel()
	w, err := em.NewWatchChannel(ctx)
	if err != nil {
		t.Fatal(err)
	}

	updates := <-w
	if len(updates) != 2 {
		t.Fatalf("expected two updates, got %+v", updates)
	}

	_, err = c.Txn(context.TODO()).Then(etcd.OpDelete("foo/host"), etcd.OpDelete("foo/host2")).Commit()
	if err != nil {
		t.Fatal(err)
	}

	updates = <-w
	if len(updates) != 2 || (updates[0].Op != endpoints.Delete && updates[1].Op != endpoints.Delete) {
		t.Fatalf("expected two delete updates, got %+v", updates)
	}
}
