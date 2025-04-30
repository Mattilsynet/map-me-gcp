package nats

import (
	"github.com/Mattilsynet/map-me-gcp/gen/mattilsynet/map-kv/key-value-watcher"
	"github.com/Mattilsynet/map-me-gcp/gen/mattilsynet/map-kv/types"
	"github.com/bytecodealliance/wasm-tools-go/cm"
)

type (
	KeyValue      struct{}
	KeyValueEntry struct {
		Key   string
		Value []byte
	}
)

func (js *JetStreamContext) KeyValue() (*KeyValue, error) {
	js.bucket = KeyValue{}
	return &js.bucket, nil
}

type KvWatcher func(kv *KeyValueEntry)

func (kv *KeyValue) RegisterKvWatchAll(fn KvWatcher) {
	keyvaluewatcher.Exports.WatchAll = func(keyValueEntry types.KeyValueEntry) (result cm.Result[string, struct{}, string]) {
		kve := KeyValueEntry{Key: keyValueEntry.Key, Value: keyValueEntry.Value.Slice()}
		fn(&kve)
		return cm.OK[cm.Result[string, struct{}, string]](struct{}{})
	}
}
