// Package main TODO
package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	servicea "github.com/vinicius-lino-figueiredo/pos-go-expert-desafio-4/internal/adapter/service-a"
	serviceb "github.com/vinicius-lino-figueiredo/pos-go-expert-desafio-4/internal/adapter/service-b"
	"github.com/vinicius-lino-figueiredo/pos-go-expert-desafio-4/internal/adapter/viacep"
	"github.com/vinicius-lino-figueiredo/pos-go-expert-desafio-4/internal/adapter/wttr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

const (
	addrA = ":8000"
	addrB = ":8080"
)

func initTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
	collectorURL := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if collectorURL == "" {
		collectorURL = "http://localhost:4318"
	}

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(collectorURL),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("weather-service"),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp, nil
}

func main() {
	http.DefaultClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	ag := viacep.NewAddressGetter(http.DefaultClient)
	tg := wttr.NewTemperatureGetter(http.DefaultClient)

	ctx, cancel := context.WithCancelCause(context.Background())

	tp, err := initTracer(ctx)
	if err != nil {
		log.Fatal("failed to initialize tracer:", err)
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Println("tracer shutdown error:", err)
		}
	}()

	hA := servicea.NewHandler("http://localhost:8080/temperature")

	hB := serviceb.NewHandler(ag, tg)

	serverA := http.Server{Addr: addrA, Handler: hA}

	serverB := http.Server{Addr: addrB, Handler: hB}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT)
	defer stop()

	go func() {
		if err := serverA.ListenAndServe(); err != nil {
			cancel(err)
		}
	}()

	go func() {
		if err := serverB.ListenAndServe(); err != nil {
			cancel(err)
		}
	}()

	<-ctx.Done()

	shDCtx, cnclShD := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cnclShD()

	if err := serverA.Shutdown(shDCtx); err != nil {
		log.Println("server A shutdown error:", err.Error())
	}
	if err := serverB.Shutdown(shDCtx); err != nil {
		log.Println("server B shutdown error:", err.Error())
	}

	log.Println(context.Cause(ctx))
}
