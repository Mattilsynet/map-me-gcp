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
	cronjob.RegisterCronHandler(mapMeGcpCronHandle)
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
		logger.Error("Failed to unmarshal ManagedEnvironment", "error", err)
		return
	}
	witManifest, err := manifest.ToWitManifest(managedGcpEnv)
	if err != nil {
		logger.Error("Failed to unmarshal WitManifest", "error", err)
		return
	}
	returnedWitManifest, err := cloudrunjobadmin.Update(witManifest)
	if err != nil {
		logger.Error("Failed to update/create cloudrunjob with manifest", "error", err)
		return
	}
	// INFO: after we've created/updated cloudrunjob responsible of creation of a given GCP ME we need to update manifest back to user
	returnedManifest, err := manifest.FromWitManifest(returnedWitManifest)
	if err != nil {
		logger.Error("Failed to unmarshal WitManifest", "error", err)
		return
	}
	returnedManifestAsBytes, err := managedenvironment.ToNatsMsg(returnedManifest)
	if err != nil {
		logger.Error("Failed to marshal ManagedEnvironment", "error", err)
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
