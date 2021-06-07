package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

type CustomUrl struct {
	url.URL
	Port string
}

type echoResponse struct {
	Method  string
	Uri     string
	Host    string
	Port    string
	Path    string
	Query   string
	Body    string
	Form    map[string]string
	Params  map[string]string
	Headers map[string]string
	TLS     *tls.ConnectionState
	URL     CustomUrl
}

func rootHandler(res http.ResponseWriter, req *http.Request) {
	log.Printf("Request %s %s\n", req.Method, req.RequestURI)

	response := echoResponse{
		Method:  req.Method,
		Uri:     req.RequestURI,
		Host:    req.Host,
		Query:   req.URL.Query().Encode(),
		Body:    "-",
		Form:    make(map[string]string),
		Headers: make(map[string]string),
		URL:     CustomUrl{*req.URL, req.URL.Port()},
		TLS:     req.TLS,
	}

	var body []byte

	req.Body.Read(body)
	response.Body = string(body)

	for name, values := range req.Form {
		for _, value := range values {
			response.Form[name] = value
		}
	}

	for name, values := range req.Header {
		for _, value := range values {
			response.Headers[name] = value
		}
	}

	res.Header().Add("Content-Type", "application/json")
	text, _ := json.MarshalIndent(response, "", "  ")
	fmt.Fprintf(res, "%s", text)
}

func hiHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(res, "<h1>Hello you!</h1>")
}

func main() {
	var listen string
	var envFlag string
	var portFlag int

	flag.StringVar(&envFlag, "env", "prod", "Environment (dev, prod)")
	flag.IntVar(&portFlag, "port", 8080, "Server port")
	flag.Parse()

	portEnv := os.Getenv("PORT")
	if portEnv == "" {
		portEnv = "8080"
	}

	if envFlag == "dev" {
		listen = "localhost:" + portEnv
	} else {
		listen = ":" + portEnv
	}

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/hi", hiHandler)

	log.Printf("Starting server on %s\n", listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}
