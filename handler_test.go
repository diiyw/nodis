package nodis

import (
	"bytes"
	"testing"

	"github.com/diiyw/nodis/redis"
)

func TestClient_SetName(t *testing.T) {
	n := &Nodis{}
	w := redis.NewWriter(&bytes.Buffer{})
	cmd := redis.Command{
		Args: []string{"SETNAME", "test"},
	}

	client(n, &redis.Conn{Writer: w}, cmd)

	expected := []byte("+OK\r\n")
	if string(w.Bytes()) != string(expected) {
		t.Errorf("Expected %q, but got %q", expected, w.Bytes())
	}
}

func TestClient_InvalidSubcommand(t *testing.T) {
	n := &Nodis{}
	w := redis.NewWriter(&bytes.Buffer{})
	cmd := redis.Command{
		Args: []string{"INVALID"},
	}

	client(n, &redis.Conn{Writer: w}, cmd)

	expected := []byte("-CLIENT subcommand must be provided\r\n")
	if string(w.Bytes()) != string(expected) {
		t.Errorf("Expected %q, but got %q", expected, w.Bytes())
	}
}

func TestClient_Config(t *testing.T) {
	n := &Nodis{}
	w := redis.NewWriter(&bytes.Buffer{})
	cmd := redis.Command{
		Name: "CONFIG",
		Args: []string{"GET", "DATABASES"},
	}

	config(n, &redis.Conn{Writer: w}, cmd)

	expected := []byte("*2\r\n$9\r\ndatabases\r\n$1\r\n0\r\n")
	if string(w.Bytes()) != string(expected) {
		t.Errorf("Expected %q, but got %q", expected, w.Bytes())
	}
}

func TestClient_Config_InvalidArgs(t *testing.T) {
	n := &Nodis{}
	w := redis.NewWriter(&bytes.Buffer{})
	cmd := redis.Command{
		Name: "CONFIG",
		Args: []string{"GET"},
	}

	config(n, &redis.Conn{Writer: w}, cmd)

	expected := []byte("-CONFIG GET requires at least two argument\r\n")
	if string(w.Bytes()) != string(expected) {
		t.Errorf("Expected %q, but got %q", expected, w.Bytes())
	}
}

func TestClient_Config_InvalidOption(t *testing.T) {
	n := &Nodis{}
	w := redis.NewWriter(&bytes.Buffer{})
	cmd := redis.Command{
		Name: "CONFIG",
		Args: []string{"GET", "INVALID"},
	}

	config(n, &redis.Conn{Writer: w}, cmd)

	expected := []byte("$-1\r\n")
	if string(w.Bytes()) != string(expected) {
		t.Errorf("Expected %q, but got %q", expected, w.Bytes())
	}
}
func TestClient_Ping(t *testing.T) {
	n := &Nodis{}
	w := redis.NewWriter(&bytes.Buffer{})
	cmd := redis.Command{
		Args: []string{"PING"},
	}

	ping(n, &redis.Conn{Writer: w}, cmd)

	expected := []byte("$4\r\nPING\r\n")
	if string(w.Bytes()) != string(expected) {
		t.Errorf("Expected %q, but got %q", expected, w.Bytes())
	}
}

func TestClient_Ping_NoArgs(t *testing.T) {
	n := &Nodis{}
	w := redis.NewWriter(&bytes.Buffer{})
	cmd := redis.Command{
		Args: []string{},
	}

	ping(n, &redis.Conn{Writer: w}, cmd)

	expected := []byte("+OK\r\n")
	if string(w.Bytes()) != string(expected) {
		t.Errorf("Expected %q, but got %q", expected, w.Bytes())
	}
}

func TestClient_Echo(t *testing.T) {
	n := &Nodis{}
	w := redis.NewWriter(&bytes.Buffer{})
	cmd := redis.Command{
		Name: "ECHO",
		Args: []string{"Hello, World!"},
	}

	echo(n, &redis.Conn{Writer: w}, cmd)

	expected := []byte("$13\r\nHello, World!\r\n")
	if string(w.Bytes()) != string(expected) {
		t.Errorf("Expected %q, but got %q", expected, w.Bytes())
	}
}

func TestClient_Echo_NoArgs(t *testing.T) {
	n := &Nodis{}
	w := redis.NewWriter(&bytes.Buffer{})
	cmd := redis.Command{
		Args: []string{},
	}

	echo(n, &redis.Conn{Writer: w}, cmd)

	expected := []byte("$-1\r\n")
	if string(w.Bytes()) != string(expected) {
		t.Errorf("Expected %q, but got %q", expected, w.Bytes())
	}
}
