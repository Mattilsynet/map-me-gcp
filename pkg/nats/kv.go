package nats

import (
	"errors"

	"github.com/Mattilsynet/map-me-gcp/gen/mattilsynet/map-kv/key-value"
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

func (js *KeyValue) Get(key string) (*KeyValueEntry, error) {
	result := keyvalue.Get(key)
	if result.IsOK() {
		resVal := result.OK().Value.Slice()
		resKey := result.OK().Key
		return &KeyValueEntry{resKey, resVal}, nil
	}
	if result.IsErr() {
		return nil, errors.New(*result.Err())
	}
	return nil, errors.New("unknown error when getting keyvalue from map-kv with key: " + key)
}

func (js *KeyValue) GetAll() ([]*KeyValueEntry, error) {
	listKeys := keyvalue.ListKeys()
	if listKeys.IsOK() {
		keys := listKeys.OK().Slice()
		var entries []*KeyValueEntry
		for _, key := range keys {
			result := keyvalue.Get(key)
			if result.IsOK() {
				resVal := result.OK().Value.Slice()
				resKey := result.OK().Key
				entries = append(entries, &KeyValueEntry{resKey, resVal})
			}
			if result.IsErr() {
				return nil, errors.New(*result.Err())
			}
		}
		return entries, nil
	}
	if listKeys.IsErr() {
		return nil, errors.New(*listKeys.Err())
	}
	return nil, errors.New("unknown error when getting all keyvalues from map-kv")
}

func (js *KeyValue) Put(key string, value []byte) error {
	result := keyvalue.Put(key, cm.ToList(value))
	if result.IsOK() {
		return nil
	}
	if result.IsErr() {
		return errors.New(*result.Err())
	}
	return errors.New("unknown error when putting keyvalue in map-kv with key: " + key)
}

func (js *KeyValue) Create(key string, value []byte) error {
	result := keyvalue.Create(key, cm.ToList(value))
	if result.IsOK() {
		return nil
	}
	if result.IsErr() {
		return errors.New(*result.Err())
	}
	return errors.New("unknown error when creating keyvalue in map-kv with key: " + key)
}

func (js *KeyValue) Delete(key string) error {
	result := keyvalue.Delete(key)
	if result.IsOK() {
		return nil
	}
	if result.IsErr() {
		return errors.New(*result.Err())
	}
	return errors.New("unknown error when deleting keyvalue in map-kv with key: " + key)
}

func (js *KeyValue) ListKeys() ([]string, error) {
	result := keyvalue.ListKeys()
	if result.IsOK() {
		return result.OK().Slice(), nil
	}
	if result.IsErr() {
		return nil, errors.New(*result.Err())
	}
	return nil, errors.New("unknown error when listing keys in map-kv")
}
