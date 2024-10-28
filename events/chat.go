package events

import (
	"time"
	log "webs/pkg/logger"
	"webs/ws"

	"go.uber.org/zap"
)

// //////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// CHAT MESSAGE
type ChatMessageIn struct {
	From    string `json:"from"`
	Message string `json:"message"`
}

type ChatMessageOut struct {
	ChatMessageIn
	Sent time.Time `json:"sent"`
}

func ChatMessageHandler(hub *ws.Hub, client *ws.Client, event *ws.Event) {
	var messageIn ChatMessageIn
	if err := event.UnmarshalPayload(&messageIn); err != nil {
		log.ZLogger.Error("[ERR-MAIN] Unmarshal messageIn", zap.Any("Payload", event.Payload))
	}

	var messageOut ChatMessageOut
	messageOut.From = messageIn.From
	messageOut.Message = messageIn.Message
	messageOut.Sent = time.Now()

	// Erstelle und sende die Message
	ws.CreateAndSendEvent(hub, "chat:message", messageOut)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// CHAT UPDATE
// TODO
