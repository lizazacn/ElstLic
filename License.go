package main

import (
	"github.com/lizazacn/ElstLic/Client"
	"github.com/lizazacn/ElstLic/Server"
)

func NewClient(offSet, step int, devInfo string) *Client.Client {
	return &Client.Client{
		Offset:  offSet,
		Step:    step,
		DevInfo: devInfo,
	}
}

func NewServer(offSet, step int, devInfo string) *Server.Server {
	return &Server.Server{
		Offset:  offSet,
		Step:    step,
		DevInfo: devInfo,
	}
}
