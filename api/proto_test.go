package api

import (
	"bytes"
	"testing"
)

func TestNewMessage(t *testing.T) {
	dto := ErrorDTO{"Oh hai der!"}
	msg, err := NewMessage(dto)

	if err != nil {
		t.Fail()
	}

	if msg.Metadata["result"] != "success" || msg.Body == nil {
		t.Fail()
	}

	body := []byte(msg.Body)
	expected := []byte(`{"message":"Oh hai der!"}`)

	if bytes.Compare(body, expected) != 0 {
		t.Fail()
	}
}

func TestNewBytesMessage(t *testing.T) {
	msg := NewBytesMessage([]byte(`{"a":"b"}`))

	if msg.Metadata["result"] != "success" || msg.Body == nil {
		t.Fail()
	}

	body := []byte(msg.Body)
	expected := []byte(`{"a":"b"}`)

	if bytes.Compare(body, expected) != 0 {
		t.Fail()
	}
}

func TestNewErrorMessage(t *testing.T) {
	msg := NewErrorMessage("Alarm!")

	if msg.Metadata["result"] != "error" || msg.Body == nil {
		t.Fail()
	}

	body := []byte(msg.Body)
	expected := []byte(`{"message":"Alarm!"}`)

	if bytes.Compare(body, expected) != 0 {
		t.Fail()
	}
}
