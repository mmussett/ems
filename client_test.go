package ems

import (
	"testing"

	"time"
)

func TestNewClient(t *testing.T) {

	ops := NewClientOptions().SetServerUrl("tcp://127.0.0.1:7222").SetUsername("admin").SetPassword("")

	c := NewClient(ops).(*client)

	if c == nil {
		t.Fatalf("ops is nil")
	}

	if c.options.serverUrl.Host != "127.0.0.1:7222" {
		t.Fatalf("bad server host")
	}

	if c.options.serverUrl.Scheme != "tcp" {
		t.Fatalf("bad server scheme")
	}

	if c.options.username != "admin" {
		t.Fatalf("bad username")
	}

	if c.options.password != "" {
		t.Fatalf("bad password")
	}

}

func TestClient_Connect(t *testing.T) {

	ops := NewClientOptions().SetServerUrl("tcp://127.0.0.1:7222").SetUsername("admin").SetPassword("")

	c := NewClient(ops).(*client)

	err := c.Connect()
	if err != nil {
		t.Fatalf(err.Error())
	}

	time.Sleep(30000 * time.Millisecond)

	c.Disconnect()

}

func TestClient_Send(t *testing.T) {

	ops := NewClientOptions().SetServerUrl("tcp://127.0.0.1:7222").SetUsername("admin").SetPassword("")

	c := NewClient(ops).(*client)

	err := c.Connect()
	if err != nil {
		t.Fatalf(err.Error())
	}

	err = c.Send("queue.sample", "hello, world", 0, TIBEMS_NON_PERSISTENT, 10000)
	if err != nil {
		t.Fatalf(err.Error())
	}

	err = c.Disconnect()
	if err != nil {
		t.Fatalf(err.Error())
	}
}
