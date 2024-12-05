package drs

import (
	"crypto/tls"
	"net/http"
)

var dc *drsClient

type drsClient struct {
	baseURL string
	token   string
	client  *http.Client
}

func init() {
	dc = &drsClient{
		//baseURL: viper.GetString("dbRemoteService"),
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}
}
