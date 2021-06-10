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
	usage = `USAGE
    echo --env ENV    # Start the server on environment ENV
    echo --help       # Shows this help

`
)

func logger(next http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 		defer func() {
		// 			r := recover()
		// 			if r != nil {
		// 				var err error
		// 				switch t := r.(type) {
		// 				case string:
		// 					err = errors.New(t)
		// 				case error:
		// 					err = t
		// 				default:
		// 					err = errors.New("Unknown error")
		// 				}
		// 				sendMeMail(err)
		// 				http.Error(w, err.Error(), http.StatusInternalServerError)
		// 			}
		// 		}()
		// 		// h.ServeHTTP(w, r)
		// 		status := newStatusResponseWriter(res)
		// 		next.ServeHTTP(status, req)

		// 	})
		// })

		// defer func() {
		// 	if r := recover(); r != nil {
		// 		fmt.Println("Recovered in f", r)
		// 		// find out exactly what the error was and set err
		// 		var err error
		// 		switch value := r.(type) {
		// 		case string:
		// 			err = errors.New(value)
		// 		case error:
		// 			err = value
		// 		default:
		// 			err = errors.New("unknown panic")
		// 		}
		// 		if err != nil {
		// 			fmt.Printf("Error report: %s.\n", err)
		// 		}
		// 	}
		// }()

		started := time.Now()

		status := newStatusResponseWriter(res)
		next.ServeHTTP(status, req)

		duration := time.Since(started).Nanoseconds()
		elapsed := fmt.Sprintf("%d ns", duration)

		if duration > 5_000_000_000 {
			elapsed = fmt.Sprintf("%0.1f sec", float64(duration)/1_000_000_000)
		} else if duration > 1_000_000 {
			elapsed = fmt.Sprintf("%0.1f ms", float64(duration)/1_000_000)
		} else if duration > 1000 {
			elapsed = fmt.Sprintf("%d µs", duration/1_000)
		}

		// Instead of hard-coding 200 capture the status code, use a wrapper
		// around http.ResponseWriter. Check:
		//
		//   https://gist.github.com/Boerworz/b683e46ae0761056a636
		//
		colorStatus := "cyan"
		if status.statusCode >= 400 {
			colorStatus = "red"
		}

		log.Printf("%s %s %s %s\n",
			color.Paint(colorStatus, status.statusCode),
			color.Green(req.Method),
			color.Green(req.URL.Path),
			color.Gray(elapsed))
	}
}

func crashHandler(res http.ResponseWriter, req *http.Request) {
	// time.Sleep(time.Nanosecond * 2_345)
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

	reply := echoReply{
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
