package web_gui

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/imightbuyaboat/SOCKS5-Proxy/server/internal/tcp"
	"github.com/imightbuyaboat/SOCKS5-Proxy/server/internal/udp"
)

type GUI interface {
	Start() error
	Shutdown(ctx context.Context) error
}

type WebGUI struct {
	portStr     string
	temp        *template.Template
	tcpListener *tcp.TCPAssociateListener
	udpListener *udp.UDPAssociateListener
	srv         *http.Server
	cancel      context.CancelFunc
}

func NewWebGUI(port int, tcpListener *tcp.TCPAssociateListener, udpListener *udp.UDPAssociateListener) GUI {
	temp := template.Must(template.ParseFiles("template/index.html"))
	portStr := fmt.Sprintf(":%d", port)

	webGui := &WebGUI{
		portStr:     portStr,
		temp:        temp,
		tcpListener: tcpListener,
		udpListener: udpListener,
	}

	r := mux.NewRouter()
	r.HandleFunc("/", webGui.mainHandler).Methods("GET")
	r.HandleFunc("/save", webGui.saveConfigHandler).Methods("POST")
	r.HandleFunc("/start", webGui.startProxyHandler).Methods("POST")
	r.HandleFunc("/stop", webGui.stopProxyHandler).Methods("POST")
	r.HandleFunc("/logs", webGui.logsHandler).Methods("GET")
	r.HandleFunc("/clear-logs", webGui.clearLogsHandler).Methods("POST")

	fs := http.FileServer(http.Dir("./static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	webGui.srv = &http.Server{
		Addr:    portStr,
		Handler: r,
	}

	return webGui
}

func (g *WebGUI) Start() error {
	return g.srv.ListenAndServe()
}

func (g *WebGUI) Shutdown(ctx context.Context) error {
	return g.srv.Shutdown(ctx)
}
