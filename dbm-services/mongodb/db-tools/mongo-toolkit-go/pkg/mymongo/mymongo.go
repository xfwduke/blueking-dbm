// Package mymongo 常用的mongo操作
package mymongo

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoHost mongo host struct
type MongoHost struct {
	Host   string `json:"host"`
	Port   string `json:"port"`
	AuthDb string `json:"authdb"`
	User   string `json:"user"`
	Pass   string `json:"-"` // 不输出到json
	Name   string `json:"name"`
	NodeIp string `json:"nodeip"`
}

func encodeURIComponent(str string) string {
	str = url.QueryEscape(str)
	str = strings.Replace(str, "+", "%20", -1)
	return str
}

// Connect return mongo client
func Connect(host, port, user, pass, authdb string, timeout time.Duration) (*mongo.Client, error) {
	return ConnectWithDirect(host, port, user, pass, authdb, timeout, false)
}

// ConnectWithDirect return mongo client
func ConnectWithDirect(host, port, user, pass, authdb string, timeout time.Duration, direct bool) (*mongo.Client, error) {
	mongoURI := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s", user, encodeURIComponent(pass), host, port, authdb)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	opt := options.Client().ApplyURI(mongoURI)
	if direct {
		opt = opt.SetDirect(true)
	}
	return mongo.Connect(ctx, opt)
}

// NewMongoHost return MongoHost
func NewMongoHost(host, port, authdb, user, pass, name, nodeip string) *MongoHost {
	obj := MongoHost{
		Host:   host,
		Port:   port,
		AuthDb: authdb,
		User:   user,
		Pass:   pass,
		Name:   name,
		NodeIp: nodeip,
	}

	if nodeip == "" {
		obj.NodeIp = obj.Host
	}

	return &obj
}

// Connect return mongo client
func (h *MongoHost) Connect() (*mongo.Client, error) {
	cli, err := Connect(h.Host, h.Port, h.User, h.Pass, h.AuthDb, 30*time.Second)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = cli.Ping(ctx, nil)
	return cli, err
}

// ConnectWithDirect return mongo client
func (h *MongoHost) ConnectWithDirect(direct bool) (*mongo.Client, error) {
	cli, err := ConnectWithDirect(h.Host, h.Port, h.User, h.Pass, h.AuthDb, 30*time.Second, direct)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = cli.Ping(ctx, nil)
	return cli, err
}

// Addr return host:port
func (h *MongoHost) Addr() string {
	return fmt.Sprintf("%s:%s", h.Host, h.Port)
}

// String Just for debug. xxx is not real password
func (h *MongoHost) String() string {
	if h.User == "" {
		return fmt.Sprintf("mongodb://%s:%s", h.Host, h.Port)
	}
	return fmt.Sprintf("mongodb://%s:%s@%s:%s/%s", h.User, "xxx", h.Host, h.Port, h.AuthDb)
}
