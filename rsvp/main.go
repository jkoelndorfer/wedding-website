package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"github.com/jkoelndorfer/wedding-website/rsvp/api"
	"github.com/jkoelndorfer/wedding-website/rsvp/config"
	"github.com/jkoelndorfer/wedding-website/rsvp/db"
	"github.com/jkoelndorfer/wedding-website/rsvp/log"
)

func main() {
	cfg := config.New()
	var invRepository db.InvitationRepository = db.New(cfg)
	http.HandleFunc("/", api.RequestHandler(cfg, invRepository))

	logger := log.Logger()
	if cfg.InLambda() {
		logger.Printf("running in Lambda mode")
		lambda.Start(httpadapter.New(http.DefaultServeMux).ProxyWithContext)
	} else {
		listenAddress := cfg.ListenAddress()
		logger.Printf("running in local server mode on %s", listenAddress)
		err := http.ListenAndServe(listenAddress, nil)
		fmt.Fprintln(os.Stderr, err)
	}
}
