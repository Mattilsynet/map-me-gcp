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
	"github.com/Mattilsynet/mapis/gen/go/managedgcpenvironment/v1"
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
		mapManagedGcpEnvironmentCronjobHandle(kve)
	}
}

/*
OBS! make sure its not recursively feeding ourselves
1. get cloudrunjob execution status
2. put execution status in mapis manifest (EXECUTION_PENDING på første, EXECUTION_SUCCEEDED / på cronjob)
3. write to kv
*/
func RemoveResourceVersion(meGcp *managedgcpenvironment.ManagedGcpEnvironment) error {
	if meGcp.Status == nil {
		meGcp.Status = &managedgcpenvironment.ManagedGcpEnvironmentStatus{}
	}
	if meGcp.Status.StatusMap == nil {
		meGcp.Status.StatusMap = make(map[string]string)
	}
	meGcp.Status.StatusMap["resource-version"] = ""
	return nil
}

func mapManagedGcpEnvironmentCronjobHandle(kve *nats.KeyValueEntry) {
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
	getWitManifest, err := cloudrunjobadmin.Get(witManifest)
	if err != nil {
		logger.Error("Failed to get cloudrunjob with manifest", "error", err)
		RemoveResourceVersion(managedGcpEnv)
		managednevironmentBytes, err := managedenvironment.ToBytes(managedGcpEnv)
		if err != nil {
			logger.Error("Failed to marshal ManagedEnvironment for gcp", "error", err)
			return
		}
		err = mapMeGcpKV.Put(kve.Key, managednevironmentBytes)
		if err != nil {
			logger.Error("Failed to put KeyValue entry", "error", err)
		}
		return
	}
	getManifest, err := manifest.FromWitManifest(getWitManifest)
	if err != nil {
		logger.Error("Failed to unmarshal WitManifest from get", "error", err)
		return
	}
	logger.Info("crj manifest status is: ", "status", getManifest.Status.GetStatusMap())
	err = manifest.AddResourceVersion(getManifest)
	if err != nil {
		logger.Error("Failed to add resource version to updated manifest", "error", err)
		return
	}
	getManifestAsBytes, err := managedenvironment.ToBytes(getManifest)
	if err != nil {
		logger.Error("Failed to marshal ManagedEnvironment for gcp", "error", err)
		return
	}
	err = mapMeGcpKV.Put(kve.Key, getManifestAsBytes)
	if err != nil {
		logger.Error("Failed to put KeyValue entry", "error", err)
		return
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
	logger.Info("manifest.isChanged pre", "kv key", kve.Key)
	isChanged := manifest.IsChanged(managedGcpEnv)
	if !isChanged {
		logger.Info("manifest.isChanged inside", "is changed", "false")
		return
	}
	logger.Info("manifest.isChanged post", "is changed", "true")
	witManifest, err := manifest.ToWitManifest(managedGcpEnv)
	if err != nil {
		logger.Error("Failed to unmarshal WitManifest", "error", err)
		return
	}
	// 1. update success
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
	err = mapMeGcpKV.Put(kve.Key, returnedManifestAsBytes)
	if err != nil {
		logger.Error("Failed to put KeyValue entry", "error", err)
		return
	}
	logger.Info("Done", "done", "yes")
}

// main should never be used in a wasm component, everything inside init()
func main() {}
