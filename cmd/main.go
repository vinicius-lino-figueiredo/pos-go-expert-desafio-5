// Package main TODO
package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	servicea "github.com/vinicius-lino-figueiredo/pos-go-expert-desafio-4/internal/adapter/service-a"
	"github.com/vinicius-lino-figueiredo/pos-go-expert-desafio-4/internal/adapter/service-b"
	"github.com/vinicius-lino-figueiredo/pos-go-expert-desafio-4/internal/adapter/viacep"
	"github.com/vinicius-lino-figueiredo/pos-go-expert-desafio-4/internal/adapter/wttr"
)

const (
	addrA = ":8000"
	addrB = ":8080"
)

func main() {
	ag := viacep.NewAddressGetter(http.DefaultClient)
	tg := wttr.NewTemperatureGetter(http.DefaultClient)

	http.DefaultClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	ctx, cancel := context.WithCancelCause(context.Background())

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
