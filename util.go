package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jpedro/color"
)

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

type jsonResponseWriter struct {
	http.ResponseWriter
}

func newStatusResponseWriter(w http.ResponseWriter) *statusResponseWriter {
	return &statusResponseWriter{w, http.StatusOK}
}

func (writer *statusResponseWriter) WriteHeader(code int) {
	writer.statusCode = code
	writer.ResponseWriter.WriteHeader(code)
}

func newJsonResponseWriter(w http.ResponseWriter) *jsonResponseWriter {
	return &jsonResponseWriter{w}
}

func (res jsonResponseWriter) send(data interface{}) {
	res.Header().Add("Content-Type", "application/json")
	text, _ := json.MarshalIndent(data, "", "  ")
	fmt.Fprintf(res, "%s", text)
}

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

		status := newStatusResponseWriter(newJsonResponseWriter(res))
		next.ServeHTTP(status, req)

		duration := time.Since(started).Nanoseconds()
		elapsed := fmt.Sprintf("%d ns", duration)

		if duration > 5e9 {
			elapsed = fmt.Sprintf("%0.1f sec", float64(duration)/1e9)
		} else if duration > 1e6 {
			elapsed = fmt.Sprintf("%0.1f ms", float64(duration)/1e6)
		} else if duration > 1e3 {
			elapsed = fmt.Sprintf("%d µs", duration/1e3)
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
