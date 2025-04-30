package nats

func (c *Conn) Jetstream() (*JetStreamContext, error) {
	return &c.js, nil
}
