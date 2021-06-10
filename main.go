package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/jpedro/color"
)

const (
	usage = `USAGE
    echo --env ENV    # Start the server on environment ENV
    echo --help       # Shows this help

`
)

type echo struct {
	Method  string            `json:"method"`
	Proto   string            `json:"protocol"`
	Host    string            `json:"host"`
	Port    string            `json:"port"`
	Uri     string            `json:"uri"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
	Path    string            `json:"path"`
	Query   string            `json:"query"`
	Params  map[string]string `json:"params"`
}

func crashHandler(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(400)
	fmt.Fprintf(res, "<h1>Crash</h1>\n")
	fmt.Fprintf(res, "Some men just like to see the world burning...\n")
}

func panicHandler(res http.ResponseWriter, req *http.Request) {
	panic("Catch me if you can")
}

func envHandler(res http.ResponseWriter, req *http.Request) {
	env := splitEnv()
	sendJson(res, env)
}

func systemHandler(res http.ResponseWriter, req *http.Request) {
	data := struct {
		OS struct {
			Release string `json:"release"`
		} `json:"os"`
		App struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"app"`
	}{}
	release, _ := ioutil.ReadFile("/etc/os-release")
	data.OS.Release = string(release)
	data.App.Name = "echo"
	data.App.Version = "v0.1.4"

	sendJson(res, data)
}

func rootHandler(res http.ResponseWriter, req *http.Request) {
	host, port := split(req.Host, ":")
	if host == "" {
		host = req.Host
	}
	if port == "" {
		port = "80 (defaulted)"
	}

	reply := echo{
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
		log.Fatalf(color.Red("ERROR Failed to parse body: %s\n", err))
	}
	reply.Body = string(body)

	for name, values := range req.Header {
		for _, value := range values {
			reply.Headers[name] = value
		}
	}

	sendJson(res, reply)
}

func main() {
	var listen string
	var envFlag string
	var portFlag int
	var helpFlag bool

	flag.BoolVar(&helpFlag, "help", false, "Shows this help")
	flag.StringVar(&envFlag, "env", "prod", "Environment (local, prod)")
	flag.IntVar(&portFlag, "port", 8080, "Server port")
	flag.Parse()

	if helpFlag {
		fmt.Print(usage)
		return
	}

	portEnv := os.Getenv("PORT")
	if portEnv == "" {
		portEnv = "8080"
	}

	if envFlag == "local" {
		listen = "localhost:" + portEnv
	} else {
		listen = ":" + portEnv
	}

	http.HandleFunc("/", logger(rootHandler))
	http.HandleFunc("/env", logger(envHandler))
	http.HandleFunc("/system", logger(systemHandler))
	http.HandleFunc("/crash", logger(crashHandler))
	http.HandleFunc("/panic", logger(panicHandler))

	log.Printf("Using env %s\n", color.Green(envFlag))
	log.Printf("Starting server on %s\n", color.Green("http://"+listen))
	log.Fatal(http.ListenAndServe(listen, nil))
}
