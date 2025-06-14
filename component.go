//go:generate go run github.com/bytecodealliance/wasm-tools-go/cmd/wit-bindgen-go generate --world map-me-gcp --out gen ./wit
package main

import (
	"log/slog"

	"go.wasmcloud.dev/component/log/wasilog"

	managedenvironment "github.com/Mattilsynet/map-me-gcp-cloudrunjob/component/pkg/managed-environment"
	"github.com/Mattilsynet/map-me-gcp-cloudrunjob/component/pkg/manifest"
	cloudrunjobadmin "github.com/Mattilsynet/map-me-gcp/pkg/cloudrunjob-admin"
	"github.com/Mattilsynet/map-me-gcp/pkg/cronjob"
	"github.com/Mattilsynet/map-me-gcp/pkg/nats"
)

var (
	logger     *slog.Logger
	conn       *nats.Conn
	mapMeGcpKV *nats.KeyValue
)

func init() {
	logger = wasilog.ContextLogger("mapMeGcp")
	logger.Info("Initializing mapMeGcp component")
	conn := nats.NewConn()
	js, err := conn.Jetstream()
	if err != nil {
		logger.Error("Failed to create Jetstream context", "error", err)
		return
	}
	mapMeGcpKV, err = js.KeyValue()
	if err != nil {
		logger.Error("Failed to create KeyValue context", "error", err)
		return
	}
	cronjob.RegisterCronHandler(mapMeGcpCronHandle)
	mapMeGcpKV.RegisterKvWatchAll(mapMeGcpHandle)
}

func mapMeGcpCronHandle() {
	logger.Info("Cron job triggered")
	kves, err := mapMeGcpKV.GetAll()
	if err != nil {
		logger.Error("Failed to get all KeyValue entries", "error", err)
		return
	}
	for _, kve := range kves {
		logger.Info("Processing KeyValue entry", "key", kve.Key)
		mapMeGcpHandle(kve)
	}
}

func mapMeGcpHandle(kve *nats.KeyValueEntry) {
	logger.Info("Handling KeyValue entry", "key", kve.Key)
	managedGcpEnvAsBytes := kve.Value
	managedGcpEnv, err := managedenvironment.ToManagedEnvironment(managedGcpEnvAsBytes)
	if err != nil {
		logger.Error("Failed to unmarshal ManagedEnvironment for gcp", "error", err)
		return
	}
	witManifest, err := manifest.ToWitManifest(managedGcpEnv)
	if err != nil {
		logger.Error("Failed to unmarshal WitManifest", "error", err)
		return
	}
	// INFO: Need to change this such that we guarantee that the cloudrunjob is up and not just check towards local state, maybe let the cloudrunjob provider handle it
	// But we also need to check this though
	// INFO: Need to update cloudrunjob provider to also yield result state of latest run

	returnedGetManifest, err := cloudrunjobadmin.Get(witManifest)
	// TODO: actually use this later
	_ = returnedGetManifest
	if err != nil {
		// INFO: Should we assume that the crj has been manually deleted on GCP here?
		logger.Error("Failed to get cloudrunjob with manifest", "error", err)
		return
	}
	// INFO: Check result of previous, also check if manifests are the same as we got after is changed
	// INFO: Put returnedGetManifest in here for validation (in nats towards manifest and towards gcp crj)
	changed := manifest.IsChanged(managedGcpEnv) // INFO: need to change cloudrunjob provider's component pkg
	if !changed {
		logger.Info("Manifest unchanged since last reconciliation: ", "key", kve.Key)
		// don't handle
		return
	}
	returnedWitManifest, err := cloudrunjobadmin.Update(witManifest)
	if err != nil {
		logger.Error("Failed to update/create cloudrunjob with manifest", "error", err)
		return
	}
	// INFO: after we've created/updated cloudrunjob responsible of creation of a given GCP ME we need to update manifest back to user
	// We can do this by giving a ready true
	returnedManifest, err := manifest.FromWitManifest(returnedWitManifest)
	if err != nil {
		logger.Error("Failed to unmarshal WitManifest", "error", err)
		return
	}
	err = manifest.AddResourceVersion(returnedManifest)
	if err != nil {
		logger.Error("Failed to add resource version to updated manifest", "error", err)
		return
	}
	returnedManifestAsBytes, err := managedenvironment.ToBytes(returnedManifest)
	if err != nil {
		logger.Error("Failed to marshal ManagedEnvironment for gcp", "error", err)
		return
	}
	// INFO: updating KV with new statuses
	err = mapMeGcpKV.Put(kve.Key, returnedManifestAsBytes)
	if err != nil {
		logger.Error("Failed to put KeyValue entry", "error", err)
		return
	}
}

// main should never be used in a wasm component, everything inside init()
func main() {}
