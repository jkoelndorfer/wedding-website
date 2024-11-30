package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"github.com/jkoelndorfer/wedding-website/rsvp/api"
	"github.com/jkoelndorfer/wedding-website/rsvp/log"
)

func main() {
	http.HandleFunc("/", api.RequestHandler)

	logger := log.Logger()
	_, in_lambda := os.LookupEnv("LAMBDA_TASK_ROOT")
	if in_lambda {
		logger.Printf("running in Lambda mode")
		lambda.Start(httpadapter.New(http.DefaultServeMux).ProxyWithContext)
	} else {
		listenAddress := "127.0.0.1:8000"
		logger.Printf("running in local server mode on %s", listenAddress)
		err := http.ListenAndServe(listenAddress, nil)
		fmt.Fprintln(os.Stderr, err)
	}
}
