package domain

import "encoding/json"

type EventType string

type Event struct {
	Provider Provider        `json:"provider,omitempty"`
	Type     EventType       `json:"event"`
	Data     json.RawMessage `json:"data,omitempty"`
}

var Events = struct {
	SessionStart EventType
	SessionEnd   EventType
	Stop         EventType
	Notification EventType
}{
	SessionStart: "session_start",
	SessionEnd:   "session_end",
	Stop:         "stop",
	Notification: "notification",
}

var providerEvents = map[Provider]map[EventType]bool{
	Providers.Claude: {
		Events.SessionStart: true,
		Events.SessionEnd:   true,
		Events.Stop:         true,
		Events.Notification: true,
	},
	Providers.Gemini: {
		Events.SessionStart: true,
		Events.SessionEnd:   true,
		Events.Stop:         true,
		Events.Notification: true,
	},
}

func (e EventType) IsValid() bool {
	switch e {
	case Events.SessionStart, Events.SessionEnd, Events.Stop, Events.Notification:
		return true
	}
	return false
}

func (e EventType) IsValidFor(p Provider) bool {
	events, ok := providerEvents[p]
	if !ok {
		return false
	}
	return events[e]
}

func (p EventType) String() string {
	return string(p)
}
