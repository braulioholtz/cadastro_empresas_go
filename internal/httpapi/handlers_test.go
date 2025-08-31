package httpapi

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"matriz/internal/models"
)

type fakeRepo struct{}

func (f *fakeRepo) Create(ctx context.Context, e *models.Empresa) (string, error) { return "1", nil }
func (f *fakeRepo) Get(ctx context.Context, id string) (*models.Empresa, error) {
	return &models.Empresa{ID: id}, nil
}
func (f *fakeRepo) GetByCNPJ(ctx context.Context, cnpj string) (*models.Empresa, error) {
	return nil, nil
}
func (f *fakeRepo) List(ctx context.Context) ([]models.Empresa, error) {
	return []models.Empresa{}, nil
}
func (f *fakeRepo) Update(ctx context.Context, id string, e *models.Empresa) error { return nil }
func (f *fakeRepo) Delete(ctx context.Context, id string) error                    { return nil }

type nopPub struct{}

func (n *nopPub) Publish(msg string) error { return nil }

func TestCreateValidation(t *testing.T) {
	api := NewServer(&fakeRepo{}, nil)
	form := url.Values{}
	form.Set("nome_fantasia", "X")
	req := httptest.NewRequest(http.MethodPost, "/empresas", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	api.create(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateFormURLEncoded(t *testing.T) {
	api := NewServer(&fakeRepo{}, nil)
	form := url.Values{}
	form.Set("cnpj", "12345678000199")
	form.Set("nome_fantasia", "Loja X")
	req := httptest.NewRequest(http.MethodPost, "/empresas", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	api.create(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestUpdateFormURLEncoded(t *testing.T) {
	api := NewServer(&fakeRepo{}, nil)
	form := url.Values{}
	form.Set("cnpj", "12345678000199")
	form.Set("nome_fantasia", "Loja X")
	req := httptest.NewRequest(http.MethodPut, "/empresas/1", strings.NewReader(form.Encode()))
	req = req.WithContext(context.Background())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	api.update(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestCreateMultipart(t *testing.T) {
	api := NewServer(&fakeRepo{}, nil)
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("cnpj", "12345678000199")
	_ = w.WriteField("nome_fantasia", "Loja Y")
	w.Close()
	req := httptest.NewRequest(http.MethodPost, "/empresas", bytes.NewReader(buf.Bytes()))
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	api.create(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}
