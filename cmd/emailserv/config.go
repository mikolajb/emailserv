package main

import "flag"

type configuration struct {
	port          string
	clientTimeout int
	amazon        struct {
		key    string
		secret string
	}
	sendgrid struct {
		key string
	}
	nop bool
}

func (c *configuration) init() {
	if c == nil {
		*c = configuration{}
	}

	flag.StringVar(&c.amazon.key, "amazon.key", "", "Amazon access key id.")
	flag.StringVar(&c.amazon.secret, "amazon.secret", "", "Amazon secret access key.")
	flag.StringVar(&c.sendgrid.key, "sendgrid.key", "", "Sendgrid key.")
	flag.IntVar(&c.clientTimeout, "client_timeout", 5000, "Acceptable client work time in milliseconds.")
	flag.StringVar(&c.port, "port", "8080", "Port.")
	flag.BoolVar(&c.nop, "nop", false, "Do not use any real client, just log messages.")
}

func (c *configuration) parse() {
	if !flag.Parsed() {
		flag.Parse()
	}
}
