package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
  "tax": 14000,
  "taxLevel": [
    {
      "level": "0-150,000",
      "tax": 0
    },
    {
      "level": "150,001-500,000",
      "tax": 14000
    },
    {
      "level": "500,001-1,000,000",
      "tax": 0
    },
    {
      "level": "1,000,001-2,000,000",
      "tax": 0
    },
    {
      "level": "2,000,001 ขึ้นไป",
      "tax": 0
    }
  ]
}`))
}

// func TestMakehttp(t *testing.T) {
// 	t.Run("Happy server response", func(t *testing.T) {
// 		e := echo.New()
// 		req := httptest.NewRequest(http.MethodGet, "/", nil)
// 		rec := httptest.NewRecorder()
// 		c := e.NewContext(req, rec)
// 		c.SetPath("/tax/calculations")
// 	})
// 	server := httptest.NewServer(http.HandlerFunc(handler))
// 	defer server.Close()
// }

func TestCalculate(t *testing.T) {
	req, err := http.NewRequest("POST", "/tax/cal", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{
  "tax": 14000,
  "taxLevel": [
    {
      "level": "0-150,000",
      "tax": 0
    },
    {
      "level": "150,001-500,000",
      "tax": 14000
    },
    {
      "level": "500,001-1,000,000",
      "tax": 0
    },
    {
      "level": "1,000,001-2,000,000",
      "tax": 0
    },
    {
      "level": "2,000,001 ขึ้นไป",
      "tax": 0
    }
  ]
}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
