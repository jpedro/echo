package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jpedro/color"
)

const (
	USAGE = `USAGE
    echo --env ENV    # Start the server on environment ENV
    echo --help       # Shows this help

`
)

func logger(next http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		now := time.Now()

		next(res, req)

		duration := time.Since(now).Nanoseconds()
		elapsed := fmt.Sprintf("%d ns", duration)

		if duration > 5_000_000_000 {
			elapsed = fmt.Sprintf("%0.1f sec", float64(duration)/1_000_000_000)
		} else if duration > 1_000_000 {
			elapsed = fmt.Sprintf("%0.1f ms", float64(duration)/1_000_000)
		} else if duration > 1000 {
			elapsed = fmt.Sprintf("%d µs", duration/1_000)
		}

		// Instead of hard-coding 200 capture the status code using a wrapper
		// around http.ResponseWriter. Check:
		//
		//   https://gist.github.com/Boerworz/b683e46ae0761056a636
		//
		log.Printf("%s %s %s %s\n",
			color.Paint("cyan", "200"),
			color.Paint("green", req.Method),
			color.Paint("green", req.URL.Path),
			color.Paint("gray", elapsed))
	}
}

func crashHandler(res http.ResponseWriter, req *http.Request) {
	// time.Sleep(time.Nanosecond * 2_345)
	res.WriteHeader(400)
	fmt.Fprintf(res, "<h1>Crash</h1>\n")
	fmt.Fprintf(res, "Some men just like to see the world burning...\n")
}

func envHandler(res http.ResponseWriter, req *http.Request) {
	sendJson(res, splitEnv())
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
		fmt.Print(USAGE)
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

	log.Printf("Using env %s\n", color.Paint("green", envFlag))
	log.Printf("Starting server on %s\n", color.Paint("green", "http://"+listen))
	log.Fatal(http.ListenAndServe(listen, nil))
}
