package adapter

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func netPipe(t *testing.T) (net.Conn, net.Conn) {
	t.Helper()
	serverConn, clientConn := net.Pipe()
	deadline := time.Now().Add(time.Second)
	if err := serverConn.SetDeadline(deadline); err != nil {
		t.Fatalf("set server deadline: %v", err)
	}
	if err := clientConn.SetDeadline(deadline); err != nil {
		t.Fatalf("set client deadline: %v", err)
	}
	t.Cleanup(func() {
		_ = serverConn.Close()
		_ = clientConn.Close()
	})
	return serverConn, clientConn
}

func errUnexpectedPayload(got, want []byte) error {
	return fmt.Errorf("payload = %q, want %q", got, want)
}

func TestTcpAdapterRoundTrip(t *testing.T) {
	serverConn, clientConn := netPipe(t)
	server, err := NewTcpAdapter(serverConn)
	if err != nil {
		t.Fatalf("NewTcpAdapter(server): %v", err)
	}
	client, err := NewTcpAdapter(clientConn)
	if err != nil {
		t.Fatalf("NewTcpAdapter(client): %v", err)
	}

	payload := []byte(`{"event":"im.message.publish","content":"hello"}`)
	serverRead := make(chan error, 1)
	go func() {
		data, err := server.Read()
		if err == nil && !bytes.Equal(data, payload) {
			err = errUnexpectedPayload(data, payload)
		}
		if err == nil {
			err = server.Write([]byte("ack"))
		}
		serverRead <- err
	}()

	if err := client.Write(payload); err != nil {
		t.Fatalf("client.Write: %v", err)
	}
	data, err := client.Read()
	if err != nil {
		t.Fatalf("client.Read: %v", err)
	}
	if string(data) != "ack" {
		t.Fatalf("client response = %q, want ack", data)
	}
	if err := <-serverRead; err != nil {
		t.Fatal(err)
	}
}

func TestTcpAdapterCloseHandler(t *testing.T) {
	serverConn, clientConn := netPipe(t)
	server, err := NewTcpAdapter(serverConn)
	if err != nil {
		t.Fatalf("NewTcpAdapter: %v", err)
	}
	closed := make(chan struct{}, 1)
	server.SetCloseHandler(func(code int, text string) error {
		if code != 1000 || text != "客户端已关闭" {
			t.Fatalf("close handler = (%d, %q)", code, text)
		}
		closed <- struct{}{}
		return nil
	})

	if err := clientConn.Close(); err != nil {
		t.Fatalf("client close: %v", err)
	}
	_, err = server.Read()
	if err == nil || !strings.Contains(err.Error(), "连接已断开") {
		t.Fatalf("Read error = %v, want connection closed error", err)
	}
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("close handler was not called")
	}
}
