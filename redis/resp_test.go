package redis

import (
	"strings"
	"testing"
)

func TestReadInlineSimple(t *testing.T) {
	doc := "ping\r\n"
	r := NewReader(strings.NewReader(doc))
	err := r.ReadCommand()
	if err != nil {
		t.Error(err)
	}
	if r.cmd.Name != "PING" {
		t.Errorf("not equal %s", r.cmd.Name)
	}
}

func TestReadInlineMutli(t *testing.T) {
	doc := "ping\r\nset foo bar\r\n"
	r := NewReader(strings.NewReader(doc))
	err := r.ReadCommand()
	if err != nil {
		t.Error(err)
	}
	if r.cmd.Name != "PING" {
		t.Errorf("not equal %s", r.cmd.Name)
	}
	err = r.ReadCommand()
	if err != nil {
		t.Error(err)
	}
	if r.cmd.Name != "SET" {
		t.Errorf("not equal %s", r.cmd.Name)
	}
	if r.cmd.Args[0] != "foo" {
		t.Errorf("not equal %s", r.cmd.Args[0])
	}
	if r.cmd.Args[1] != "bar" {
		t.Errorf("not equal '%s'", r.cmd.Args[1])
	}
}

func TestReadInlineQuates(t *testing.T) {
	doc := "set \"foo\" bar\r\nset foo \"bar\"\r\n\r\nset foo \"bar bar2\"\r\n"
	r := NewReader(strings.NewReader(doc))
	err := r.ReadCommand()
	if err != nil {
		t.Error(err)
	}
	if r.cmd.Name != "SET" {
		t.Errorf("not equal %s", r.cmd.Name)
	}
	if r.cmd.Args[0] != "foo" {
		t.Errorf("not equal %s", r.cmd.Args[0])
	}
	if r.cmd.Args[1] != "bar" {
		t.Errorf("not equal '%s'", r.cmd.Args[1])
	}
	err = r.ReadCommand()
	if err != nil {
		t.Error(err)
	}
	if r.cmd.Name != "SET" {
		t.Errorf("not equal %s", r.cmd.Name)
	}
	if r.cmd.Args[0] != "foo" {
		t.Errorf("not equal %s", r.cmd.Args[0])
	}
	if r.cmd.Args[1] != "bar" {
		t.Errorf("not equal '%s'", r.cmd.Args[1])
	}
}

func TestReadInlineSpace(t *testing.T) {
	doc := "set foo \"bar bar2\"\r\nset  foo \"bar\"bar2\"\r\n"
	r := NewReader(strings.NewReader(doc))
	err := r.ReadCommand()
	if err != nil {
		t.Error(err)
	}
	if r.cmd.Name != "SET" {
		t.Errorf("not equal %s", r.cmd.Name)
	}
	if r.cmd.Args[0] != "foo" {
		t.Errorf("not equal %s", r.cmd.Args[0])
	}
	if r.cmd.Args[1] != "bar bar2" {
		t.Errorf("not equal '%s'", r.cmd.Args[1])
	}
	err = r.ReadCommand()
	if err != nil {
		t.Error(err)
	}
	if r.cmd.Name != "SET" {
		t.Errorf("not equal %s", r.cmd.Name)
	}
	if r.cmd.Args[0] != "foo" {
		t.Errorf("not equal %s", r.cmd.Args[0])
	}
	if r.cmd.Args[1] != "bar" {
		t.Errorf("not equal '%s'", r.cmd.Args[1])
	}
}

func TestInlineQuotes(t *testing.T) {
	doc := "set  foo \"bar\\\"bar3\"\r\n"
	r := NewReader(strings.NewReader(doc))
	err := r.ReadCommand()
	if err != nil {
		t.Error(err)
	}
	if r.cmd.Name != "SET" {
		t.Errorf("not equal %s", r.cmd.Name)
	}
	if r.cmd.Args[0] != "foo" {
		t.Errorf("not equal %s", r.cmd.Args[0])
	}
	if r.cmd.Args[1] != "bar\\\"bar3" {
		t.Errorf("not equal '%s'", r.cmd.Args[1])
	}
}
