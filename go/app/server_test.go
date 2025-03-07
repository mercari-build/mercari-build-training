package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"
)

func TestParseAddItemRequest(t *testing.T) {
	t.Parallel()

	type wants struct {
		req *AddItemRequest
		err bool
	}

	cases := map[string]struct {
		args map[string]string
		wants
	}{
		"ok: valid request": {
			args: map[string]string{
				"name":     "Test Item",
				"category": "Electronics",
			},
			wants: wants{
				req: &AddItemRequest{
					Name:     "Test Item",
					Category: "Electronics",
				},
				err: false,
			},
		},
		"ng: empty request": {
			args: map[string]string{},
			wants: wants{
				req: nil,
				err: true,
			},
		},
		"ng: missing name": {
			args: map[string]string{
				"category": "Electronics",
			},
			wants: wants{
				req: nil,
				err: true,
			},
		},
		"ng: missing category": {
			args: map[string]string{
				"name": "Test Item",
			},
			wants: wants{
				req: nil,
				err: true,
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			values := url.Values{}
			for k, v := range tt.args {
				values.Set(k, v)
			}

			req, err := http.NewRequest("POST", "http://localhost:9000/items", strings.NewReader(values.Encode()))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			got, err := parseAddItemRequest(req)

			if err != nil {
				if !tt.err {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if diff := cmp.Diff(tt.wants.req, got); diff != "" {
				t.Errorf("unexpected request (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAddItem(t *testing.T) {
	t.Parallel()

	type wants struct {
		code int
	}
	cases := map[string]struct {
		args     map[string]string
		injector func(m *MockItemRepository)
		wants
	}{
		"ok: correctly inserted": {
			args: map[string]string{
				"name":     "Used MacBook",
				"category": "Laptop",
			},
			injector: func(m *MockItemRepository) {
				m.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(nil)
			},
			wants: wants{
				code: http.StatusCreated,
			},
		},
		"ng: failed to insert": {
			args: map[string]string{
				"name":     "Used MacBook",
				"category": "Laptop",
			},
			injector: func(m *MockItemRepository) {
				m.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(errors.New("insert failed"))
			},
			wants: wants{
				code: http.StatusInternalServerError,
			},
		},
		"ng: missing name": {
			args: map[string]string{
				"category": "Laptop",
			},
			injector: func(m *MockItemRepository) {},
			wants: wants{
				code: http.StatusBadRequest,
			},
		},
		"ng: missing category": {
			args: map[string]string{
				"name": "Used MacBook",
			},
			injector: func(m *MockItemRepository) {},
			wants: wants{
				code: http.StatusBadRequest,
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockIR := NewMockItemRepository(ctrl)
			tt.injector(mockIR)

			h := &Handlers{itemRepo: mockIR}

			values := url.Values{}
			for k, v := range tt.args {
				values.Set(k, v)
			}
			req := httptest.NewRequest("POST", "/items", strings.NewReader(values.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rr := httptest.NewRecorder()
			h.AddItem(rr, req)

			if tt.wants.code != rr.Code {
				t.Errorf("expected status code %d, got %d", tt.wants.code, rr.Code)
			}
		})
	}
}

func TestAddItemHandler(t *testing.T) {
	t.Parallel()

	type wants struct {
		code int
		body string
	}

	cases := map[string]struct {
		requestBody map[string]string
		injector    func(m *MockItemRepository)
		wants
	}{
		"ok: correctly inserted": {
			requestBody: map[string]string{
				"name":     "Gaming Laptop",
				"category": "Electronics",
			},
			injector: func(m *MockItemRepository) {
				m.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(nil)
			},
			wants: wants{
				code: http.StatusCreated,
				body: `{"message":"item added"}`,
			},
		},
		"ng: missing name": {
			requestBody: map[string]string{
				"category": "Electronics",
			},
			injector: func(m *MockItemRepository) {},
			wants: wants{
				code: http.StatusBadRequest,
				body: "name is required",
			},
		},
		"ng: missing category": {
			requestBody: map[string]string{
				"name": "Gaming Laptop",
			},
			injector: func(m *MockItemRepository) {},
			wants: wants{
				code: http.StatusBadRequest,
				body: "category is required",
			},
		},
		"ng: insert failed": {
			requestBody: map[string]string{
				"name":     "Gaming Laptop",
				"category": "Electronics",
			},
			injector: func(m *MockItemRepository) {
				m.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(errors.New("database error"))
			},
			wants: wants{
				code: http.StatusInternalServerError,
				body: "failed to save item",
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := NewMockItemRepository(ctrl)
			tt.injector(mockRepo)

			h := &Handlers{itemRepo: mockRepo}

			// JSON エンコードされたリクエストボディ
			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/items", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			h.AddItem(rr, req)

			// ステータスコードの検証
			if rr.Code != tt.wants.code {
				t.Errorf("expected status %d, got %d", tt.wants.code, rr.Code)
			}

			// レスポンスボディの検証
			if tt.wants.body != "" && !bytes.Contains(rr.Body.Bytes(), []byte(tt.wants.body)) {
				t.Errorf("expected response body to contain %q, got %q", tt.wants.body, rr.Body.String())
			}
		})
	}
}
