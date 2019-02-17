package api

import (
	"encoding/json"
	"testing"
)

func TestBuildTextMessage(t *testing.T) {
	b := Builder()
	b.Header("a", "b")
	b.Text("Hello world!")

	msg := b.Message()

	if msg == nil {
		t.Error("Message was nil")
	}

	if msg.Metadata["a"] != "b" {
		t.Errorf("Header [a] was not b, but: %s.", msg.Metadata["a"])
	}

	body, err := msg.String()

	if err != nil {
		t.Errorf("Did not expect an error reading body as a string: %s", err)
	}

	if body != "Hello world!" {
		t.Errorf("Body was not Hello world! but: %s.", body)
	}

	_, err = json.Marshal(msg)

	if err != nil {
		t.Errorf("Message could not be serialized: %s", err)
	}
}

func TestBuildJsonMessage(t *testing.T) {
	b := Builder()
	b.Json(&ErrorDTO{"Hello world!"})

	msg := b.Message()

	if msg == nil {
		t.Error("Message was nil")
	}

	dto := &ErrorDTO{}
	err := msg.Json(dto)

	if err != nil {
		t.Errorf("Did not expect a json error: %s", err.Error())
	}

	if dto.Message != "Hello world!" {
		t.Errorf("Message was not Hello world! but: %s.", dto.Message)
	}
}
