package nats

import (
	"errors"

	"github.com/Mattilsynet/map-me-gcp/gen/wasmcloud/messaging/consumer"
	"github.com/Mattilsynet/map-me-gcp/gen/wasmcloud/messaging/types"
	"github.com/bytecodealliance/wasm-tools-go/cm"
)

type (
	Conn struct {
		js JetStreamContext
	}
	JetStreamContext struct {
		bucket KeyValue
	}
	Msg struct {
		Subject string
		Reply   string
		Data    []byte
		Header  map[string][]string
	}
)

func NewConn() *Conn {
	return &Conn{}
}

func ToBrokenMessageFromNatsMessage(nm *Msg) types.BrokerMessage {
	if nm.Reply == "" {
		return types.BrokerMessage{
			Subject: nm.Subject,
			Body:    cm.ToList(nm.Data),
			ReplyTo: cm.None[string](),
		}
	} else {
		return types.BrokerMessage{
			Subject: nm.Subject,
			Body:    cm.ToList(nm.Data),
			ReplyTo: cm.Some(nm.Subject),
		}
	}
}

func (nc *Conn) Publish(msg *Msg) error {
	bm := ToBrokenMessageFromNatsMessage(msg)
	result := consumer.Publish(bm)
	if result.IsErr() {
		return errors.New(*result.Err())
	}
	return nil
}
