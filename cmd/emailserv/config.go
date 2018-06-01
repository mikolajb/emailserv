package main

import "flag"

type configuration struct {
	amazon struct {
		key    string
		secret string
	}
	sendgrid struct {
		key string
	}
	clientTimeout int
}

func (c *configuration) init() {
	if c == nil {
		*c = configuration{}
	}

	flag.StringVar(&c.amazon.key, "amazon.key", "", "amazon access key id")
	flag.StringVar(&c.amazon.secret, "amazon.secret", "", "amazon secret access key")
	flag.StringVar(&c.sendgrid.key, "sendgrid.key", "", "sendgrid key")
	flag.IntVar(&c.clientTimeout, "client_timeout", 5000, "acceptable client work time in milliseconds")
}

func (c *configuration) parse() {
	if !flag.Parsed() {
		flag.Parse()
	}
}
