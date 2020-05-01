package handlers

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (hS HttpServer) ServeAnsibleOutput(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hS.Logger.Print(r.URL.Path)
	hS.Logger.Print("Request to serve logs/ansible_output.log")
	p := "./logs/ansible_output.log"
	http.ServeFile(w, r, p)
}

func (hS HttpServer) ServeAnsibleServiceLog(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hS.Logger.Print(r.URL.Path)
	hS.Logger.Print("Request to serve logs/launcher.log")
	p := "./logs/launcher.log"
	http.ServeFile(w, r, p)
}

func (hS HttpServer) ServeHttpServerLog(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hS.Logger.Print(r.URL.Path)
	hS.Logger.Print("Request to serve logs/http_server.log")
	p := "./logs/http_server.log"
	http.ServeFile(w, r, p)
}
