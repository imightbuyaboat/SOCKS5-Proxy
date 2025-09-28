package web_gui

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/config"
)

type GuiConfig struct {
	TCPRelayServerAddress string `json:"tcp_relay_server_address"`
	UDPRelayServerAddress string `json:"udp_relay_server_address"`
}

func (g *WebGUI) getCurrentConfig() *GuiConfig {
	return &GuiConfig{
		TCPRelayServerAddress: g.listenerTCP.GetAddress(),
		UDPRelayServerAddress: g.listenerUDP.GetAddress(),
	}
}

func (g *WebGUI) mainHandler(w http.ResponseWriter, r *http.Request) {
	guiConfig := g.getCurrentConfig()

	if err := g.temp.ExecuteTemplate(w, "index.html", guiConfig); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (g *WebGUI) saveConfigHandler(w http.ResponseWriter, r *http.Request) {
	var newConfig GuiConfig

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		response := map[string]interface{}{
			"error":          "invalid json",
			"current_config": g.getCurrentConfig(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	if err := config.ValidateAddress(newConfig.TCPRelayServerAddress); err != nil {
		response := map[string]interface{}{
			"error":          "invalid tcp relay address",
			"current_config": g.getCurrentConfig(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	if err := config.ValidateAddress(newConfig.UDPRelayServerAddress); err != nil {
		response := map[string]interface{}{
			"error":          "invalid udp relay address",
			"current_config": g.getCurrentConfig(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	g.listenerTCP.UpdateAddress(newConfig.TCPRelayServerAddress)
	g.listenerUDP.UpdateAddress(newConfig.UDPRelayServerAddress)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"config updated"}`))
}

func (g *WebGUI) startProxyHandler(w http.ResponseWriter, r *http.Request) {
	if g.cancel != nil {
		http.Error(w, `{"error":"relay already running"}`, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	g.cancel = cancel

	go func() {
		g.listenerTCP.Start(ctx)
	}()
	go func() {
		g.listenerUDP.Start(ctx)
	}()

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"relay started"}`))
}

func (g *WebGUI) stopProxyHandler(w http.ResponseWriter, r *http.Request) {
	if g.cancel == nil {
		http.Error(w, `{"error":"relay not running"}`, http.StatusBadRequest)
		return
	}

	g.cancel()
	g.cancel = nil

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"relay stopped"}`))
}

func (g *WebGUI) logsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(
		map[string]string{"logs": g.listenerTCP.GetLogs()},
	); err != nil {
		http.Error(w, `{"error":"error while getting logs"}`, http.StatusInternalServerError)
		return
	}
}

func (g *WebGUI) clearLogsHandler(w http.ResponseWriter, r *http.Request) {
	g.listenerTCP.ClearLogs()
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"logs cleared"}`))
}
