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

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatalf("Error loading .env file: %s", err)
		}

		// Getting and using a value from .env
		username := os.Getenv("ADMIN_USERNAME")

		fmt.Println(username)
		return c.String(http.StatusOK, "Hello, Go Bootcamp!")
	})
	go func() {
		e.Logger.Fatal(e.Start(":1323"))
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
