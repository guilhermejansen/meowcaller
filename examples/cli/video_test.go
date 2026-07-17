package main

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVideoBridgeControlDispatchesJSONCommand(t *testing.T) {
	vb := &videoBridge{}
	var got vbControl
	vb.OnControl(func(command vbControl) error {
		got = command
		return nil
	})
	req := httptest.NewRequest(http.MethodPost, "/control", bytes.NewBufferString(`{"action":"dial_video","target":"15551234567"}`))
	rec := httptest.NewRecorder()

	vb.handleControl(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if got.Action != "dial_video" || got.Target != "15551234567" {
		t.Fatalf("control = %+v", got)
	}
}

func TestVideoBridgeControlReportsCommandFailure(t *testing.T) {
	vb := &videoBridge{}
	vb.OnControl(func(vbControl) error { return errors.New("no active call") })
	req := httptest.NewRequest(http.MethodPost, "/control", bytes.NewBufferString(`{"action":"hangup"}`))
	rec := httptest.NewRecorder()

	vb.handleControl(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", rec.Code)
	}
}

func TestVideoBridgeControlRejectsUnknownAction(t *testing.T) {
	vb := &videoBridge{}
	vb.OnControl(func(vbControl) error { return nil })
	req := httptest.NewRequest(http.MethodPost, "/control", bytes.NewBufferString(`{"action":"explode"}`))
	rec := httptest.NewRecorder()

	vb.handleControl(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
