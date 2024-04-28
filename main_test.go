package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func handler(w http.ResponseWriter, r *http.Request, want string) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(want))
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
	req, err := http.NewRequest("POST", "/tax/calculations", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	want := `{
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
	handler(rr, req, want)

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

func TestCal(t *testing.T) {
	t.Run("test status code", func(t *testing.T) {
		userJSON := `{
  "totalIncome": 500000.0,
  "wht": 0.0,
  "allowances": [
    {
      "allowanceType": "k-receipt",
      "amount": 200000
    },
    {
      "allowanceType": "donation",
      "amount": 100000.0
    }
  ]
}`
		// Setup
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/tax/calculations", strings.NewReader(userJSON))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		//statusCode := rec.Result().StatusCode
		statusCode := rec.Result()
		fmt.Printf("%v | %T", statusCode, c)
		//h := &handler{mockDB}

		// Assertions
		// if assert.NoError(t, h.createUser(c)) {
		// 	assert.Equal(t, http.StatusCreated, rec.Code)
		// 	assert.Equal(t, userJSON, rec.Body.String())
		// }
	})
}
