package api

import "testing"

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

	body := string(msg.Body)

	if body != "Hello world!" {
		t.Errorf("Body was not Hello world! but: %s.", body)
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
