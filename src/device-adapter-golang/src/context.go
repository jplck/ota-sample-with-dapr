package main

import "os"

type Context struct {
	Host     string
	Port     string
	ClientID string
}

func (ctx Context) Init() Context {
	ctx.ClientID, _ = os.LookupEnv("CLIENT_ID")
	ctx.Port, _ = os.LookupEnv("IOT_HUB_PORT")
	ctx.Host, _ = os.LookupEnv("IOT_HUB_HOST")
	return ctx
}

func (ctx Context) Valid() bool {
	return ctx.ClientID != "" && ctx.Host != "" && ctx.Port != ""
}
