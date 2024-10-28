package main

import (
	"fmt"
	"net/http"
	"webs/config"
	"webs/events"
	log "webs/pkg/logger"
	"webs/ws"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// main is the entry point of the application.
func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		panic(fmt.Errorf("config error: %+v", err))
	}

	// Logger
	zapLogger := log.NewLogger(cfg.Log.Level, cfg.Log.OutputPath, cfg.Log.ErrOutputPath)
	defer zapLogger.Sync()

	// Router
	router := chi.NewRouter()

	// Hub
	hub := ws.NewHub(&cfg.Hub)
	hub.CreateRooms([]string{
		"chat",
	}, &cfg.Room)

	// Static files
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/chat.html")
	})

	// WebSocket
	router.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		// Register a new client.
		client, err := hub.OnConnect(w, r, &cfg.Client)
		if err != nil {
			log.ZLogger.Fatal("OnConnect error", zap.Error(err))
		}

		// Register event handlers.
		eventHandlers := map[string]func(*ws.Hub, *ws.Client, *ws.Event){
			"subscribe:chat":     events.SubscribeHandler,
			"subscribe:roulette": events.SubscribeHandler,

			"unsubscribe:chat":     events.UnsubscribeHandler,
			"unsubscribe:roulette": events.UnsubscribeHandler,

			"chat:message": events.ChatMessageHandler,
		}

		// Execute event handlers.
		for event, handler := range eventHandlers {
			hub.On(event, func(event *ws.Event) {
				handler(hub, client, event)
			})
		}
	})

	// Start the web server.
	log.ZLogger.Info("Webserver is listening on port", zap.String("Port", cfg.App.Port))
	err = http.ListenAndServe(":"+cfg.App.Port, router)
	if err != nil {
		log.ZLogger.Fatal("ListenAndServe error", zap.Error(err))
	}
}
