package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/oklog/run"
	"github.com/opencensus-integrations/ocsql"
	zipkin "github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"go.opencensus.io/exporter/prometheus"
	oczipkin "go.opencensus.io/exporter/zipkin"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"

	"github.com/basvanbeek/ocsql-example/ocmux"
	"github.com/basvanbeek/ocsql-example/server"
)

const (
	port = "8080"
)

func main() {
	// set-up our Zipkin details
	var (
		reporter         = zipkinhttp.NewReporter("http://localhost:9411/api/v2/spans")
		localEndpoint, _ = zipkin.NewEndpoint("ocsql", "localhost:"+port)
		zipkinExporter   = oczipkin.NewExporter(reporter, localEndpoint)
	)
	defer reporter.Close()

	// Always trace for this demo.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	// use Zipkin for tracing
	trace.RegisterExporter(zipkinExporter)

	// set-up our Prometheus details
	prometheusExporter, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "ocsql",
	})
	if err != nil {
		log.Fatalf("error configuring prometheus: %v\n", err)
	}

	// add default ochttp server views
	view.Register(ochttp.DefaultServerViews...)

	// Report stats at every second.
	view.SetReportingPeriod(1 * time.Second)

	// use Prometheus for metrics
	view.RegisterExporter(prometheusExporter)

	// create our ocsql wrapper
	driverName, err := ocsql.Register("sqlite3", ocsql.WithAllTraceOptions())
	if err != nil {
		log.Fatalf("error registering ocsql: %v\n", err)
	}

	// create our ephemeral SQLite database
	db, err := sql.Open(driverName, "file::memory:?mode=memory&cache=shared")
	if err != nil {
		log.Fatalf("error opening our SQLite database: %v\n", err)
	}

	// create our SQLite Repository and prime it
	repository := server.NewSQLiteRepository(db)
	repository.Prime(context.Background())

	// create our Service
	svc := server.New(repository)

	router := mux.NewRouter()
	router.Use(ocmux.Middleware())

	// attach our service endpoint
	router.Methods("GET").Path("/user/{user_id}/items").HandlerFunc(svc.ListItems)

	// attach prometheus endpoint
	router.Handle("/metrics", prometheusExporter)

	// wrap our router with ochttp tracing and metrics
	handler := &ochttp.Handler{Handler: router}

	// add a listener for our service
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("unable to create listener: %v", err)
	}

	// run.Group manages our goroutine lifecycles
	// see: https://www.youtube.com/watch?v=LHe1Cb_Ud_M&t=15m45s
	var g run.Group
	// set-up our HTTP service handler
	{
		g.Add(func() error {
			return http.Serve(listener, handler)

		}, func(error) {
			listener.Close()
		})
	}
	// set-up our signal handler
	{
		var (
			cancelInterrupt = make(chan struct{})
			c               = make(chan os.Signal, 2)
		)
		defer close(c)

		g.Add(func() error {
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				return fmt.Errorf("received signal %s", sig)
			case <-cancelInterrupt:
				return nil
			}
		}, func(error) {
			close(cancelInterrupt)
		})
	}

	// spawn our goroutines and wait for shutdown
	log.Println("exit", g.Run())
}
