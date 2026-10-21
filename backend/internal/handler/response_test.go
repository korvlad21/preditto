package handler_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func successData(t *testing.T, w *httptest.ResponseRecorder, field string) json.RawMessage {
	t.Helper()
	var body map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body) != 2 || string(body["success"]) != "true" || body["data"] == nil ||
		w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatalf("invalid success envelope: %s", w.Body.String())
	}
	if field == "" {
		return body["data"]
	}
	var data map[string]json.RawMessage
	if err := json.Unmarshal(body["data"], &data); err != nil {
		t.Fatal(err)
	}
	if len(data) != 1 || data[field] == nil {
		t.Fatalf("missing collection %s: %s", field, w.Body.String())
	}
	return data[field]
}
