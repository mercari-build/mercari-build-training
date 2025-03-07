package app

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
)

func TestParseAddItemRequest(t *testing.T) {
	t.Parallel()

	type wants struct {
		req *AddItemRequest
		err bool
	}

	// STEP 6-1: define test cases
	cases := map[string]struct {
		args map[string]string
		wants
	}{
		"ok: valid request": {
			args: map[string]string{
				"name":     "item",    // fill here
				"category": "fashion", // fill here
			},
			wants: wants{
				req: &AddItemRequest{
					Name:     "item",    // fill here
					Category: "fashion", // fill here
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
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// prepare request body
			values := url.Values{}
			for k, v := range tt.args {
				values.Set(k, v)
			}

			// prepare HTTP request
			req, err := http.NewRequest("POST", "http://localhost:9000/items", strings.NewReader(values.Encode()))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// execute test target
			got, err := parseAddItemRequest(req)

			// confirm the result
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

func TestHelloHandler(t *testing.T) {
	t.Parallel()

	// Please comment out for STEP 6-2
	// predefine what we want
	type wants struct {
		code int               // desired HTTP status code
		body map[string]string //desired body
	}
	want := wants{
		code: http.StatusOK,
		body: map[string]string{"message": "Hello, world!"},
	}

	// set up test
	req := httptest.NewRequest("GET", "/hello", nil)
	res := httptest.NewRecorder()

	h := &Handlers{}
	h.Hello(res, req)

	// STEP 6-2: confirm the status code
	if res.Code != want.code {
		t.Errorf("want %d, but %d", want.code, res.Code)
	}
	// STEP 6-2: confirm response body
	// convert JSON to GO struct or map
	var gotBody map[string]string
	err := json.Unmarshal(res.Body.Bytes(), &gotBody)
	if err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	// Compare structures when == can't be used
	if !reflect.DeepEqual(gotBody, want.body) {
		t.Errorf("want %v, but got %v", want.body, gotBody)
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
				"name":     "used iPhone 16e",
				"category": "phone",
			},
			injector: func(m *MockItemRepository) {
				// STEP 6-3: define mock expectation
				m.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(&Item{ID: 1, Name: "used iPhone 16e", Category: "phone", Image: "default.jpg"}, nil).Times(1)
				// succeeded to insert
			},
			wants: wants{
				code: http.StatusOK,
			},
		},
		"ng: failed to insert": {
			args: map[string]string{
				"name":     "used iPhone 16e",
				"category": "phone",
			},
			injector: func(m *MockItemRepository) {
				// STEP 6-3: define mock expectation
				// gomock.Any()→ 引数はなんでもいい
				m.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(nil, errors.New("insert error")).Times(1)
				// failed to insert
			},
			wants: wants{
				code: http.StatusInternalServerError,
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockIR := NewMockItemRepository(ctrl)
			tt.injector(mockIR)
			h := &Handlers{repo: mockIR}

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
			if tt.wants.code >= 400 {
				return
			}

			for _, v := range tt.args {
				if !strings.Contains(rr.Body.String(), v) {
					t.Errorf("response body does not contain %s, got: %s", v, rr.Body.String())
				}
			}
		})
	}
}

// STEP 6-4: uncomment this test
func TestAddItemE2e(t *testing.T) {
	// create a buffer to store the multipart request
	if testing.Short() {
		t.Skip("skipping e2e test")
	}

	db, closers, err := setupDB(t)
	if err != nil {
		t.Fatalf("failed to set up database: %v", err)
	}
	t.Cleanup(func() {
		for _, c := range closers {
			c()
		}
	})

	type wants struct {
		code int
	}
	cases := map[string]struct {
		args map[string]string
		wants
	}{
		"ok: correctly inserted": {
			args: map[string]string{
				"name":     "used iPhone 16e",
				"category": "phone",
			},
			wants: wants{
				code: http.StatusOK,
			},
		},
		"ng: failed to insert": {
			args: map[string]string{
				"name":     "",
				"category": "phone",
			},
			wants: wants{
				code: http.StatusBadRequest,
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			h := &Handlers{repo: &itemRepository{db: db}}

			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)
			_ = writer.WriteField("name", tt.args["name"])
			_ = writer.WriteField("category", tt.args["category"])
			writer.Close()

			part, err := writer.CreateFormFile("image", "test.jpg")
			if err != nil {
				t.Fatalf("failed to create form file: %v", err)
			}
			part.Write([]byte("dummy image data"))

			writer.Close()

			req := httptest.NewRequest("POST", "/items", &buf)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			rr := httptest.NewRecorder()
			h.AddItem(rr, req)

			// check response
			if tt.wants.code != rr.Code {
				t.Errorf("expected status code %d, got %d", tt.wants.code, rr.Code)
			}
			if tt.wants.code >= 400 {
				return
			}

			// STEP 6-4: check inserted data
			// Get item from actual database
			item := &Item{}
			err = db.QueryRow("SELECT name, category FROM items WHERE name = ?", tt.args["name"]).Scan(&item.Name, &item.Category)
			if err != nil {
				t.Fatalf("failed to query inserted item: %v", err)
			}

			//verify that the inserted item matches the expected values
			if item.Name != tt.args["name"] {
				t.Errorf("expected item name to be %s,got %s", tt.args["name"], item.Name)
			}
			if item.Category != tt.args["category"] {
				t.Errorf("expected item category to be %s, got %s", tt.args["category"], item.Category)
			}

		})
	}
}

func setupDB(t *testing.T) (db *sql.DB, closers []func(), e error) {
	t.Helper()

	defer func() {
		if e != nil {
			for _, c := range closers {
				c()
			}
		}
	}()

	// create a temporary file for e2e testing
	f, err := os.CreateTemp(".", "*.sqlite3")
	if err != nil {
		return nil, nil, err
	}
	closers = append(closers, func() {
		f.Close()
		os.Remove(f.Name())
	})

	// set up tables
	db, err = sql.Open("sqlite3", f.Name())
	if err != nil {
		return nil, nil, err
	}
	closers = append(closers, func() {
		db.Close()
	})

	// TODO: replace it with real SQL statements.

	cmd := `CREATE TABLE IF NOT EXISTS categories(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(255)`

	_, err = db.Exec(cmd)
	if err != nil {
		return nil, nil, err
	}

	cmd = `CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(255),
		category VARCHAR(255)
	)`

	_, err = db.Exec(cmd)
	if err != nil {
		return nil, nil, err
	}

	return db, closers, nil
}
