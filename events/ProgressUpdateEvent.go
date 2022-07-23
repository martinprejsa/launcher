package events

import "fmt"

var ProgressUpdateEvent progressUpdateEvent

type ProgressUpdateEventPayload struct {
	Progress float64
}

type progressUpdateEvent struct {
	handlers []interface {
		Handle(ProgressUpdateEventPayload)
	}
}

func (p *progressUpdateEvent) Register(handler interface {
	Handle(event ProgressUpdateEventPayload)
}) {
	p.handlers = append(p.handlers, handler)
}

func (p progressUpdateEvent) Trigger(payload ProgressUpdateEventPayload) {
	fmt.Println(payload.Progress)
	for _, handler := range p.handlers {
		go handler.Handle(payload)
	}
}
