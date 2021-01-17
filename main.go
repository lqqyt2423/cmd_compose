package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/lqqyt2423/cmd_compose/compose"
)

func main() {
	log.SetOutput(os.Stdout)
	log.Printf("pid: %v\n", os.Getpid())

	var filename string
	flag.StringVar(&filename, "f", "cmd_compose.json", "config file")
	flag.Parse()

	log.Printf("use config file: %v\n", filename)
	configs, err := compose.Parse(filename)
	if err != nil {
		log.Fatalf("parse %v error: %v\n", filename, err)
	}

	ct := compose.NewController(configs)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-sigs
		log.Printf("got signal: %v\n", s)
		ct.Kill()
	}()

	ct.Run()
	log.Println("exit")
}
