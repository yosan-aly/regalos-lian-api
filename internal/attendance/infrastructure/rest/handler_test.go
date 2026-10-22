package rest

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/yosan-aly/regalos-lian-api/internal/attendance/domain"
	"github.com/yosan-aly/regalos-lian-api/internal/platform/httperr"
)

type stubService struct {
	err error
}

func (s *stubService) Confirm(_ context.Context, name string, attendees int) (domain.Confirmation, error) {
	if s.err != nil {
		return domain.Confirmation{}, s.err
	}
	return domain.Confirmation{ID: "c1", Name: name, Attendees: attendees}, nil
}

func (s *stubService) Update(_ context.Context, id, name string, attendees int) (domain.Confirmation, error) {
	if s.err != nil {
		return domain.Confirmation{}, s.err
	}
	return domain.Confirmation{ID: id, Name: name, Attendees: attendees}, nil
}

func (s *stubService) Summary(context.Context) (domain.Summary, error) {
	return domain.Summary{
		Confirmations:  []domain.Confirmation{{ID: "a", Attendees: 2}, {ID: "b", Attendees: 3}},
		TotalAttendees: 5,
	}, nil
}

func (s *stubService) Delete(context.Context, string) error { return s.err }

func newTestApp(svc AttendanceService) *fiber.App {
	app := fiber.New()
	h := NewHandler(svc)
	app.Post("/attendance-confirmations", h.Confirm)
	app.Put("/attendance-confirmations/:id", h.Update)
	app.Get("/admin/attendance-confirmations", h.Summary)
	app.Delete("/admin/attendance-confirmations/:id", h.Delete)
	return app
}

func doJSON(t *testing.T, app *fiber.App, method, path, body string) (int, []byte) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, raw
}

func TestConfirmAndUpdate_StatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		serviceErr error
		wantStatus int
		wantCode   string
	}{
		{name: "confirm created", method: "POST", path: "/attendance-confirmations", body: `{"name":"Ana","attendees":3}`, wantStatus: fiber.StatusCreated},
		{name: "confirm invalid name", method: "POST", path: "/attendance-confirmations", body: `{"name":"A","attendees":1}`, serviceErr: domain.ErrInvalidName, wantStatus: fiber.StatusBadRequest, wantCode: "INVALID_NAME"},
		{name: "confirm malformed json", method: "POST", path: "/attendance-confirmations", body: `{`, wantStatus: fiber.StatusBadRequest, wantCode: "INVALID_BODY"},
		{name: "update ok", method: "PUT", path: "/attendance-confirmations/x", body: `{"name":"Ana","attendees":2}`, wantStatus: fiber.StatusOK},
		{name: "update not found", method: "PUT", path: "/attendance-confirmations/x", body: `{"name":"Ana","attendees":2}`, serviceErr: domain.ErrConfirmationNotFound, wantStatus: fiber.StatusNotFound, wantCode: "CONFIRMATION_NOT_FOUND"},
		{name: "delete not found", method: "DELETE", path: "/admin/attendance-confirmations/x", serviceErr: domain.ErrConfirmationNotFound, wantStatus: fiber.StatusNotFound, wantCode: "CONFIRMATION_NOT_FOUND"},
		{name: "delete ok", method: "DELETE", path: "/admin/attendance-confirmations/x", wantStatus: fiber.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, raw := doJSON(t, newTestApp(&stubService{err: tt.serviceErr}), tt.method, tt.path, tt.body)
			if status != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", status, tt.wantStatus, raw)
			}
			if tt.wantCode == "" {
				return
			}
			var body httperr.ErrorResponse
			if err := json.Unmarshal(raw, &body); err != nil {
				t.Fatal(err)
			}
			if body.Error.Code != tt.wantCode {
				t.Fatalf("code = %q, want %q", body.Error.Code, tt.wantCode)
			}
		})
	}
}

func TestSummary_IncludesTotals(t *testing.T) {
	status, raw := doJSON(t, newTestApp(&stubService{}), "GET", "/admin/attendance-confirmations", "")
	if status != fiber.StatusOK {
		t.Fatalf("status = %d", status)
	}
	var body ConfirmationSummaryResponse
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	if body.TotalConfirmations != 2 || body.TotalAttendees != 5 {
		t.Fatalf("totals = %d/%d, want 2/5", body.TotalConfirmations, body.TotalAttendees)
	}
}
