package web_gui

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/socks5"
)

type GUI interface {
	Start() error
	Shutdown(ctx context.Context) error
}

type WebGUI struct {
	portStr  string
	temp     *template.Template
	listener *socks5.SOCKS5Listener
	srv      *http.Server
	cancel   context.CancelFunc
}

func NewWebGUI(port int, listener *socks5.SOCKS5Listener) GUI {
	temp := template.Must(template.ParseFiles("template/index.html"))
	portStr := fmt.Sprintf(":%d", port)

	webGUI := &WebGUI{
		portStr:  portStr,
		temp:     temp,
		listener: listener,
	}

	r := mux.NewRouter()
	r.HandleFunc("/", webGUI.mainHandler).Methods("GET")
	r.HandleFunc("/save", webGUI.saveConfigHandler).Methods("POST")
	r.HandleFunc("/start", webGUI.startProxyHandler).Methods("POST")
	r.HandleFunc("/stop", webGUI.stopProxyHandler).Methods("POST")
	r.HandleFunc("/logs", webGUI.logsHandler).Methods("GET")
	r.HandleFunc("/clear-logs", webGUI.clearLogsHandler).Methods("POST")

	fs := http.FileServer(http.Dir("./static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	webGUI.srv = &http.Server{
		Addr:    portStr,
		Handler: r,
	}

	return webGUI
}

func (g *WebGUI) Start() error {
	return g.srv.ListenAndServe()
}

func (g *WebGUI) Shutdown(ctx context.Context) error {
	return g.srv.Shutdown(ctx)
}
