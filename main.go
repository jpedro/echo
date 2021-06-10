package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jpedro/color"
)

type EchoReply struct {
	Method  string            `json:"method"`
	Proto   string            `json:"protocol"`
	Host    string            `json:"host"`
	Port    string            `json:"port"`
	Uri     string            `json:"uri"`
	Path    string            `json:"path"`
	Query   string            `json:"query"`
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers"`
	Params  map[string]string `json:"params"`
}

func envHandler(res http.ResponseWriter, req *http.Request) {
	log.Printf("Request %s %s\n", req.Method, req.RequestURI)
	sendJson(res, splitEnv())
}

func systemHandler(res http.ResponseWriter, req *http.Request) {
	log.Printf("Request %s %s\n", req.Method, req.RequestURI)
	res.Header().Add("Content-Type", "application/json")

	data := struct {
		OS struct {
			Release string
		}
		App struct {
			Name    string
			Version string
		}
	}{}
	release, _ := ioutil.ReadFile("/etc/os-release")
	data.OS.Release = string(release)

	sendJson(res, data)
}

func rootHandler(res http.ResponseWriter, req *http.Request) {
	log.Printf("Request %s %s\n", req.Method, req.RequestURI)

	host, port := split(req.Host, ":")
	if host == "" {
		host = req.Host
	}
	if port == "" {
		port = "80 (defaulted)"
	}

	reply := EchoReply{
		Method:  req.Method,
		Proto:   req.Proto,
		Host:    host,
		Port:    port,
		Uri:     req.RequestURI,
		Path:    req.URL.Path,
		Query:   req.URL.Query().Encode(),
		Params:  splitParams(req.URL.Query().Encode()),
		Body:    "-",
		Headers: make(map[string]string),
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatalf("ERROR Failed to parse body: %s\n", err)
	}
	reply.Body = string(body)

	for name, values := range req.Header {
		for _, value := range values {
			reply.Headers[name] = value
		}
	}

	sendJson(res, reply)
}

func sendJson(res http.ResponseWriter, data interface{}) {
	res.Header().Add("Content-Type", "application/json")
	text, _ := json.MarshalIndent(data, "", "  ")
	fmt.Fprintf(res, "%s", text)
}

func split(text string, separator string) (string, string) {
	index := strings.Index(text, separator)

	if index < 0 {
		return "", ""
	}

	before := text[:index]
	after := text[index+len(separator):]

	return before, after
}

func splitEnv() map[string]string {
	env := make(map[string]string)
	for _, value := range os.Environ() {
		key, val := split(value, "=")
		env[key] = val
	}

	return env
}

func splitParams(query string) map[string]string {
	res := make(map[string]string)

	if len(query) < 1 {
		return res
	}

	for _, segment := range strings.Split(query, "&") {
		key, val := split(segment, "=")
		res[key] = val
	}

	return res
}

func hiHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(res, "<h1>Hello you!</h1>")
}

func main() {
	var listen string
	var envFlag string
	var portFlag int

	flag.StringVar(&envFlag, "env", "prod", "Environment (local, prod)")
	flag.IntVar(&portFlag, "port", 8080, "Server port")
	flag.Parse()

	portEnv := os.Getenv("PORT")
	if portEnv == "" {
		portEnv = "8080"
	}

	if envFlag == "local" {
		listen = "localhost:" + portEnv
	} else {
		listen = ":" + portEnv
	}

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/hi", hiHandler)
	http.HandleFunc("/env", envHandler)
	http.HandleFunc("/system", systemHandler)

	log.Printf("Using env %s\n", color.Paint("green", envFlag))
	log.Printf("Starting server on %s\n", color.Paint("green", "http://"+listen))
	log.Fatal(http.ListenAndServe(listen, nil))
}
