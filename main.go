package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

type income struct {
	TotalIncome float32     `json:"totalIncome" validate:"required"`
	Wht         float32     `json:"wht"`
	Allowances  []allowance `json:"allowances"`
}
type allowance struct {
	AllowanceType string  `json:"allowanceType"`
	Amount        float32 `json:"amount"`
}
type tax struct {
	Tax      float32    `json:"tax"`
	TaxLevel []taxLevel `json:"taxLevel"`
}
type taxLevel struct {
	Level string  `json:"level"`
	Tax   float32 `json:"tax"`
}

func main() {
	e := echo.New()
	e.Use()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Go Bootcamp!")
	})
	e.POST("/tax/calculations", calculations, somemiddleware)
	e.POST("/admin/deductions/personal", updateDeducatePerson, AuthAdmin)
	go func() {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatalf("Error loading .env file: %s", err)
		}
		e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", os.Getenv("PORT"))))
	}()
	//graceful shutdown
	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, os.Interrupt, syscall.SIGTERM)
	//signal.Notify(gracefulStop, os.Interrupt, syscall.SIGINT)

	<-gracefulStop
	fmt.Println("Server shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		fmt.Printf("Error shutting down server %s", err)
	} else {
		fmt.Println("shut down the server")
	}

}
func somemiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		//fmt.Println("SomeMiddleware")
		/*return c.JSON(
			http.StatusBadGateway,
			map[string]any{"message": "error"},
		)*/ // if error
		return next(c)
	}
}
func AuthAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		//fmt.Println("SomeMiddleware")
		/*return c.JSON(
			http.StatusBadGateway,
			map[string]any{"message": "error"},
		)*/ // if error
		return next(c)
	}
}
func calculations(c echo.Context) error {
	inc := new(income)
	t := new(tax)
	t = &tax{Tax: 0.0}
	err := c.Bind(&inc)
	if err != nil {
		return c.JSON(http.StatusBadGateway, err)
	} else if inc.TotalIncome < 0 || inc.Wht < 0 {
		return c.JSON(http.StatusBadGateway, "Please add positive number")
	}
	l := []struct {
		textLevel string
		min       float32
		max       float32
		rate      int
	}{
		{textLevel: "0-150,000", min: 1, max: 150000, rate: 0},
		{textLevel: "150,001-500,000", min: 150001, max: 500000, rate: 10},
		{textLevel: "500,001-1,000,000", min: 500001, max: 1000000, rate: 15},
		{textLevel: "1,000,001-2,000,000", min: 1000001, max: 2000000, rate: 20},
		{textLevel: "2,000,001 ขึ้นไป", min: 2000001, max: 10000000000, rate: 35},
	}

	numTax := calDeduction(inc)
	var tempTax, temp float32
	for _, v := range l {
		if numTax >= v.min {
			if numTax > v.max {
				temp = v.max - (v.min - 1)
			} else {
				temp = (numTax - (v.min - 1))
			}
			fmt.Printf("%.1f\n", temp)
			tempTax = (temp * float32(v.rate) / 100)
			t.Tax = t.Tax + tempTax
			//fmt.Printf("%.1f | %.1f | %.1f\n", numTax, temp, t.Tax)
		} else {
			tempTax = 0
		}
		t.TaxLevel = append(t.TaxLevel, taxLevel{Level: v.textLevel, Tax: tempTax})
	}
	calWht(inc, t)
	return c.JSON(http.StatusOK, t)
}
func calDeduction(inc *income) float32 {
	m := (inc.Allowances)
	numTax := inc.TotalIncome - 60000.0
	for _, v := range m {
		//fmt.Printf("type => %s | amount => %.1f", v.AllowanceType, v.Amount)
		if v.AllowanceType == "donation" {
			if v.Amount > 100000.0 {
				numTax = numTax - 100000.0
			} else {
				numTax = numTax - v.Amount
			}
		}
	}
	return numTax
}
func calWht(inc *income, t *tax) {
	if inc.Wht > 0 {
		t.Tax = t.Tax - inc.Wht
		if t.Tax < 0 {
			t.Tax = 0
		}
	}
}
func updateDeducatePerson(c echo.Context) error {
	return c.JSON(http.StatusOK, c)
}
