package rest

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/yosan-aly/regalos-lian-api/internal/gifts/domain"
)

type stubService struct {
	reserveErr error
	gotQty     int
}

func (s *stubService) Catalog(context.Context) ([]domain.Category, error) {
	return []domain.Category{{
		ID: "paseo", Title: "Paseo",
		Gifts: []domain.Gift{{ID: "coche", Name: "Coche", Required: 1, Selected: 0}},
	}}, nil
}

func (s *stubService) Reserve(_ context.Context, giftID string, qty int) (domain.Reservation, domain.Gift, error) {
	s.gotQty = qty
	if s.reserveErr != nil {
		return domain.Reservation{}, domain.Gift{}, s.reserveErr
	}
	return domain.Reservation{ID: "r1", Quantity: qty},
		domain.Gift{ID: giftID, Name: "Coche", Required: 1, Selected: 1}, nil
}

func (s *stubService) Reservations(context.Context, bool) ([]domain.Reservation, error) {
	return nil, nil
}

func (s *stubService) CancelReservation(context.Context, string) error { return nil }

func newTestApp(svc GiftService) *fiber.App {
	app := fiber.New()
	h := NewHandler(svc)
	app.Get("/catalog", h.Catalog)
	app.Post("/gifts/:giftId/reservations", h.Reserve)
	return app
}

func TestCatalog_IncludesAvailable(t *testing.T) {
	app := newTestApp(&stubService{})

	resp, err := app.Test(httptest.NewRequest("GET", "/catalog", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body catalogResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if got := body.Categories[0].Gifts[0].Available; got != 1 {
		t.Fatalf("available = %d, want 1", got)
	}
}

func TestReserve_StatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		serviceErr error
		wantStatus int
		wantCode   string
	}{
		{name: "created", body: `{"quantity":2}`, wantStatus: fiber.StatusCreated},
		{name: "created without body", body: ``, wantStatus: fiber.StatusCreated},
		{name: "sold out", body: `{}`, serviceErr: domain.ErrNotEnoughAvailable, wantStatus: fiber.StatusConflict, wantCode: "NOT_ENOUGH_AVAILABLE"},
		{name: "unknown gift", body: `{}`, serviceErr: domain.ErrGiftNotFound, wantStatus: fiber.StatusNotFound, wantCode: "GIFT_NOT_FOUND"},
		{name: "invalid quantity", body: `{"quantity":99}`, serviceErr: domain.ErrInvalidQuantity, wantStatus: fiber.StatusBadRequest, wantCode: "INVALID_QUANTITY"},
		{name: "malformed json", body: `{`, wantStatus: fiber.StatusBadRequest, wantCode: "INVALID_BODY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApp(&stubService{reserveErr: tt.serviceErr})
			req := httptest.NewRequest("POST", "/gifts/coche/reservations", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				raw, _ := io.ReadAll(resp.Body)
				t.Fatalf("status = %d, want %d (body: %s)", resp.StatusCode, tt.wantStatus, raw)
			}
			if tt.wantCode == "" {
				return
			}
			var body errorResponse
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body.Error.Code != tt.wantCode {
				t.Fatalf("code = %q, want %q", body.Error.Code, tt.wantCode)
			}
		})
	}
}

func TestReserve_DefaultsQuantityToOne(t *testing.T) {
	svc := &stubService{}
	app := newTestApp(svc)
	req := httptest.NewRequest("POST", "/gifts/coche/reservations", strings.NewReader(``))
	req.Header.Set("Content-Type", "application/json")

	if _, err := app.Test(req); err != nil {
		t.Fatal(err)
	}
	if svc.gotQty != 1 {
		t.Fatalf("quantity = %d, want 1", svc.gotQty)
	}
}
