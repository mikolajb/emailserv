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
}

func (c *configuration) init() {
	if c == nil {
		*c = configuration{}
	}

	flag.StringVar(&c.amazon.key, "amazon.key", "", "amazon access key id")
	flag.StringVar(&c.amazon.secret, "amazon.secret", "", "amazon secret access key")
	flag.StringVar(&c.sendgrid.key, "sendgrid.key", "", "sendgrid key")
}

func (c *configuration) parse() {
	if !flag.Parsed() {
		flag.Parse()
	}
}
