package network

import (
	"net/url"
	"testing"
)

func TestUrl(t *testing.T) {
	t.Run("websocket", func(t *testing.T) {
		addr := "ws://127.0.0.1:8848"

		l, err := url.Parse(addr)
		if err != nil {
			t.Fail()
		}

		t.Log(l.Scheme)
		t.Log(l.Host)
	})

	t.Run("tcp", func(t *testing.T) {
		addr := "tcp://127.0.0.1:8848"

		l, err := url.Parse(addr)
		if err != nil {
			t.Fail()
		}

		t.Log(l.Scheme)
		t.Log(l.Host)
	})
}
