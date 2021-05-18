package main

import (
	"flag"
	"github.com/clearcodecn/binance-bot/pkg/http"
	"log"
	"os"
	"os/signal"
	"time"
)

var (
	config string
)

func init() {
	flag.StringVar(&config, "c", "config.yaml", "config file")
}

func main() {
	flag.Parse()

	server := http.NewServer(config)
	stopCh := make(chan struct{})
	go func() {
		log.Fatal(server.Run(stopCh, ":8080"))
	}()

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill)
	<-ch
	close(stopCh)
	time.Sleep(3 * time.Second)
}
