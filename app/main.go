package main

import (
	"context"
	"syscall"

	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"time"

	"github.com/ardanlabs/service/app/services/sales-api"
)

/*
Need to figure out timeouts for http service.
Use env for all configuration values.
*/

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func main() {

	// ============================================================
	// Configuration

	readTimeout := 5 * time.Second
	writeTimeout := 10 * time.Second
	shutdownTimeout := 5 * time.Second
	host := os.Getenv("HOST")
	if host == "" {
		host = ":3000"
	}

	// ============================================================
	// Start Service

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)


	server := http.Server{
		Addr:           host,
		Handler:        sales_api.APIMux(sales_api.APIMuxConfig{
			Shutdown: shutdown}),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	// Starting the service, listening for requests.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		log.Printf("startup : Listening %s", host)
		log.Printf("shutdown : Listener closed : %v", server.ListenAndServe())
		wg.Done()
	}()


	// ============================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	select {
	case sig := <- shutdown:
		log.Println("stop", sig)
		// Create context for Shutdown call.
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and load shed.
		if err := server.Shutdown(ctx); err != nil {
		log.Printf("shutdown : Graceful shutdown did not complete in %v : %v", shutdownTimeout, err)

		if err := server.Close(); err != nil {
		log.Printf("shutdown : Error killing server : %v", err)
		}
	}
}

	// Waiting for service to complete that load shedding.
	wg.Wait()
	log.Println("main : Completed")
}