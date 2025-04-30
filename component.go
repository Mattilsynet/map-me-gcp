package main

import (
	"log/slog"

	"go.wasmcloud.dev/component/log/wasilog"

	"github.com/Mattilsynet/map-me-gcp/pkg/cronjob"
	"github.com/Mattilsynet/map-me-gcp/pkg/nats"
)

var (
	logger *slog.Logger
	conn   *nats.Conn
)

func init() {
	logger = wasilog.ContextLogger("mapMeGcp")
	cronjob.RegisterCronHandler(mapMeGcpCronHandle)
	conn := nats.NewConn()
	js, err := conn.Jetstream()
	if err != nil {
		logger.Error("Failed to create Jetstream context", "error", err)
		return
	}
	kv, err := js.KeyValue()
	if err != nil {
		logger.Error("Failed to create KeyValue context", "error", err)
		return
	}
	kv.RegisterKvWatchAll(mapMeGcpWatch)
}

func mapMeGcpCronHandle() {
	logger.Info("Cronjob handler called")
	// need to do kv.GetAll and surf through
	// should align for each with what mapMeGcpWatch also uses
}

func mapMeGcpWatch(kve *nats.KeyValueEntry) {
}

// main should never be used in a wasm component, everything inside init()
func main() {}
