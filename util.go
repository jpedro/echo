package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

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
