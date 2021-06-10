package main

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
