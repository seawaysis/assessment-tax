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

type Tax struct {
	TotalIncome float32      `json:"totalIncome"`
	Wht         float32      `json:"wht"`
	Allowances  *[]allowance `json:"allowances"`
}
type allowance struct {
	AllowanceType string  `json:"allowanceType"`
	Amount        float32 `json:"amount"`
}

func main() {
	e := echo.New()
	e.Use()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Go Bootcamp!")
	})
	e.POST("/tax/calculations", calculations, somemiddleware)

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
func calculations(c echo.Context) error {
	t := new(Tax)
	err := c.Bind(&t)
	if err != nil {
		return c.JSON(http.StatusBadGateway, err)
	}
	return c.JSON(http.StatusOK, t)
	//return c.String(http.StatusOK, c.Param("totalIncome"))
}
