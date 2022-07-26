package events

var ProgressUpdateEvent progressUpdateEvent

type ProgressUpdateEventPayload struct {
	Progress float64
	Message  string
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
	for _, handler := range p.handlers {
		go handler.Handle(payload)
	}
}
