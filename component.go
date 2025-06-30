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

/*
OBS! make sure its not recursively feeding ourselves
1. update / create cloudrunjob
2. get cloudrunjob execution status
3. put execution status in mapis manifest (EXECUTION_PENDING på første, EXECUTION_SUCCEEDED / på cronjob)
4. write to kv
*/
func mapManagedGcpEnvironmentHandle(kve *nats.KeyValueEntry) {
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
	returnedWitManifest, err := cloudrunjobadmin.Update(witManifest)
	if err != nil {
		logger.Error("Failed to update/create cloudrunjob with manifest", "error", err)
		return
	}

	returnedManifest, err := manifest.FromWitManifest(returnedWitManifest)
	if err != nil {
		logger.Error("Failed to unmarshal WitManifest", "error", err)
		return
	}
	getWitManifest, err := cloudrunjobadmin.Get(returnedWitManifest)
	if err != nil {
		logger.Error("Failed to get cloudrunjob with manifest", "error", err)
		return
	}
	getManifest, err := manifest.FromWitManifest(getWitManifest)
	if err != nil {
		logger.Error("Failed to unmarshal WitManifest from get", "error", err)
		return
	}
	logger.Info("crj manifet status is: ", "status", getManifest.Status.GetStatusMap())
	// INFO: Logisk brist, denne vil alltid sette resource version i status til det samme som resource version i spec, som betyr at isChanged alltid vil returnere false
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
	// check if returnedMAnifestAsBytes == witManifest
	if manifest.IsChanged(returnedManifest) {
		err = mapMeGcpKV.Put(kve.Key, returnedManifestAsBytes)
		if err != nil {
			logger.Error("Failed to put KeyValue entry", "error", err)
			return
		}
	}
	logger.Info("Done", "done", "yes")
}

// main should never be used in a wasm component, everything inside init()
func main() {}
