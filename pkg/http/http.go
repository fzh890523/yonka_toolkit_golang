package http

import (
	"fmt"
	"flag"
)

type ServerConf struct {
	Host string // ip or host
	Port int
}

func (c *ServerConf) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *ServerConf) RegisterFlag() {
	flag.StringVar(&c.Host, "httpServerHost", "", "host or ip or empty")
	flag.IntVar(&c.Port, "httpServerPort", 9999, "host or ip or empty")
}

func (c *ServerConf) ParseFlag() error {
	return nil
}
