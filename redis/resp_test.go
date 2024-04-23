package redis

import (
	"strings"
	"testing"
)

func TestReadInlineSimple(t *testing.T) {
	doc := "ping\r\n"
	r := NewResp(strings.NewReader(doc))
	v, err := r.Read()
	if err != nil {
		t.Error(err)
	}
	if v.Array[0].Bulk != "ping" {
		t.Errorf("not equal %s", v.Bulk)
	}
}

func TestReadInlineMutli(t *testing.T) {
	doc := "ping\r\nset foo bar\r\n"
	r := NewResp(strings.NewReader(doc))
	v, err := r.Read()
	if err != nil {
		t.Error(err)
	}
	if v.Array[0].Bulk != "ping" {
		t.Errorf("not equal %s", v.Bulk)
	}
	v, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if v.Array[0].Bulk != "set" {
		t.Errorf("not equal %s", v.Array[0].Bulk)
	}
	if v.Array[1].Bulk != "foo" {
		t.Errorf("not equal %s", v.Array[1].Bulk)
	}
	if v.Array[2].Bulk != "bar" {
		t.Errorf("not equal %s", v.Array[2].Bulk)
	}
}

func TestReadInlineQuates(t *testing.T) {
	doc := "set \"foo\" bar\r\nset foo \"bar\"\r\n\r\nset foo \"bar bar2\"\r\n"
	r := NewResp(strings.NewReader(doc))
	v, err := r.Read()
	if err != nil {
		t.Error(err)
	}
	if v.Array[0].Bulk != "set" {
		t.Errorf("not equal %s", v.Array[0].Bulk)
	}
	if v.Array[1].Bulk != "foo" {
		t.Errorf("not equal %s", v.Array[1].Bulk)
	}
	if v.Array[2].Bulk != "bar" {
		t.Errorf("not equal '%s'", v.Array[2].Bulk)
	}
	v, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if v.Array[0].Bulk != "set" {
		t.Errorf("not equal %s", v.Array[0].Bulk)
	}
	if v.Array[1].Bulk != "foo" {
		t.Errorf("not equal %s", v.Array[1].Bulk)
	}
	if v.Array[2].Bulk != "bar" {
		t.Errorf("not equal '%s'", v.Array[2].Bulk)
	}
}

func TestReadInlineSpace(t *testing.T) {
	doc := "set foo \"bar bar2\"\r\nset foo \"bar\"bar2\"\r\n"
	r := NewResp(strings.NewReader(doc))
	v, err := r.Read()
	if err != nil {
		t.Error(err)
	}
	if v.Array[0].Bulk != "set" {
		t.Errorf("not equal %s", v.Array[0].Bulk)
	}
	if v.Array[1].Bulk != "foo" {
		t.Errorf("not equal %s", v.Array[1].Bulk)
	}
	if v.Array[2].Bulk != "bar bar2" {
		t.Errorf("not equal '%s'", v.Array[2].Bulk)
	}
	v, err = r.Read()
	if err != nil {
		t.Error(err)
	}
	if v.Array[0].Bulk != "set" {
		t.Errorf("not equal %s", v.Array[0].Bulk)
	}
	if v.Array[1].Bulk != "foo" {
		t.Errorf("not equal %s", v.Array[1].Bulk)
	}
	if v.Array[2].Bulk != "bar" {
		t.Errorf("not equal '%s'", v.Array[2].Bulk)
	}
}
