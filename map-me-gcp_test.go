package main

import (
	"testing"

	me_gcp "github.com/Mattilsynet/mapis/gen/go/managedgcpenvironment/v1"
	metadata "github.com/Mattilsynet/mapis/gen/go/meta/v1"
	"github.com/nats-io/nats.go"
)

func TestMapMeUpdate(t *testing.T) {
	me := &me_gcp.ManagedGcpEnvironment{}
	spec := me_gcp.ManagedGcpEnvironmentSpec{}
	oMeta := &metadata.ObjectMeta{}
	oMeta.Name = "test-test2"
	me.Metadata = oMeta
	me.Metadata.ResourceVersion = "1"
	spec.BudgetAmount = "100"
	spec.DnsZoneName = "DZone"
	spec.TeamArRepoId = "Super-repo"
	spec.Email = "superduper@super-mail.com"
	spec.Group = "group2"
	spec.MapspaceRef = "some-team-name-2"
	spec.ParentFolderId = "123123124"
	me.Spec = &spec
	t.Errorf("%s", me.Metadata.Name)
	meAsBytes, err := me.MarshalVT()
	if err != nil {
		t.Fail()
	}
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		t.Errorf("Failed to connect to NATS: %v", err)
	}
	js, err := nc.JetStream()
	if err != nil {
		t.Errorf("Failed to create JetStream context: %v", err)
	}
	kv, err := js.KeyValue("map-me-gcp")
	if err != nil {
		t.Errorf("Failed to create KeyValue store: %v", err)
	}
	kv.Put(spec.MapspaceRef, meAsBytes)
}
