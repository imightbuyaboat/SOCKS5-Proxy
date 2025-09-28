package web_gui

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/imightbuyaboat/SOCKS5-Proxy/server/internal/tcp"
	"github.com/imightbuyaboat/SOCKS5-Proxy/server/internal/udp"
)

type WebGUI struct {
	port int

	temp *template.Template

	listenerTCP *tcp.TCPAssociateListener
	listenerUDP *udp.UDPAssociateListener

	cancel context.CancelFunc
}

func NewWebGUI(port int, listenerTCP *tcp.TCPAssociateListener, listenerUDP *udp.UDPAssociateListener) *WebGUI {
	temp := template.Must(template.ParseFiles("template/index.html"))

	return &WebGUI{
		port:        port,
		temp:        temp,
		listenerTCP: listenerTCP,
		listenerUDP: listenerUDP,
	}
}

func (g *WebGUI) Start() {
	portStr := fmt.Sprintf(":%d", g.port)

	r := mux.NewRouter()
	r.HandleFunc("/", g.mainHandler).Methods("GET")
	r.HandleFunc("/save", g.saveConfigHandler).Methods("POST")
	r.HandleFunc("/start", g.startProxyHandler).Methods("POST")
	r.HandleFunc("/stop", g.stopProxyHandler).Methods("POST")
	r.HandleFunc("/logs", g.logsHandler).Methods("GET")
	r.HandleFunc("/clear-logs", g.clearLogsHandler).Methods("POST")

	fs := http.FileServer(http.Dir("./static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	if err := http.ListenAndServe(portStr, r); err != nil {
		log.Fatal(err)
	}
}
