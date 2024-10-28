package ws

import (
	"encoding/json"
	"errors"
	"strings"
)

// EventHandler defines a function type for handling events.
type EventHandler func(*Event)

// Event represents a WebSocket event with a type and an optional payload.
type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// NewEventFromRaw creates an Event from a raw byte message.
func NewEventFromRaw(message []byte) (*Event, error) {
	var rawMessage []interface{}
	// Unmarshal the raw byte message into a slice of interfaces.
	if err := json.Unmarshal(message, &rawMessage); err != nil || len(rawMessage) == 0 {
		return nil, errors.New("unmarshal Event-Format")
	}

	// Extract the event type, which must be a string.
	eventType, ok := rawMessage[0].(string)
	if !ok {
		return nil, errors.New("Event-Typ muss ein String sein")
	}

	// Create the event and set the payload if available.
	event := &Event{
		Type: eventType,
	}
	if len(rawMessage) > 1 {
		event.Payload = rawMessage[1]
	}

	return event, nil
}

// Raw converts the Event into a byte slice.
func (event *Event) Raw() ([]byte, error) {
	rawArray := []interface{}{event.Type}
	// Append the payload if it exists.
	if event.Payload != nil {
		rawArray = append(rawArray, event.Payload)
	}
	return json.Marshal(rawArray)
}

// UnmarshalPayload unmarshals the Event's payload into the given target.
func (event *Event) UnmarshalPayload(target interface{}) error {
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(payloadBytes, target)
}

// CreateAndSendEvent creates an Event and broadcasts it to the specified room.
func CreateAndSendEvent(hub *Hub, eventType string, payload interface{}) {
	outgoingEvent := &Event{
		Type:    eventType,
		Payload: payload,
	}

	// Determine the room from the event type and broadcast the event.
	room := strings.Split(eventType, ":")[0]
	hub.BroadcastToRoom(room, outgoingEvent)
}

// CreateAndSendEventToClient creates an Event and sends it to the specified client.
func CreateAndSendEventToClient(hub *Hub, client *Client, eventType string, payload interface{}) {
	outgoingEvent := &Event{
		Type:    eventType,
		Payload: payload,
	}

	// Send the event directly to the client.
	hub.SendToClient(client, outgoingEvent)
}
