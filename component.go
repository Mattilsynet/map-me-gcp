//go:generate go run github.com/bytecodealliance/wasm-tools-go/cmd/wit-bindgen-go generate -world mapMeGcp -out gen ./wit
package main

import (
	"log/slog"

	"go.wasmcloud.dev/component/log/wasilog"

	keyvaluewatcher "github.com/Mattilsynet/map-me-gcp/gen/mattilsynet/map-kv/key-value-watcher"
	"github.com/Mattilsynet/map-me-gcp/gen/mattilsynet/map-kv/types"
	"github.com/Mattilsynet/map-me-gcp/pkg/cronjob"
	"github.com/bytecodealliance/wasm-tools-go/cm"
)

var logger *slog.Logger

func init() {
	logger = wasilog.ContextLogger("mapMeGcp")
	cronjob.RegisterCronHandler(mapMeGcpCronHandler)
	keyvaluewatcher.Exports.WatchAll = mapMeGcpHandler
}

func mapMeGcpCronHandler() {
	logger.Info("Cronjob handler called")
}

func mapMeGcpHandler(kve types.KeyValueEntry) cm.Result[string, struct{}, string] {
	return cm.OK[cm.Result[string, struct{}, string]](struct{}{})
}

// main should never be used in a wasm component, everything inside init()
func main() {}
