package eventbus

type Event struct {
	Type    string
	Payload map[string]interface{}
}

type Subscriber func(Event)

var subscribers []Subscriber

func Subscribe(sub Subscriber) {
	subscribers = append(subscribers, sub)
}

func Publish(event Event) {
	for _, sub := range subscribers {
		sub(event)
	}
	logEvent(event)
}

func logEvent(event Event) {
	// Use logger if available, else fallback to println
	// logger.Info("eventbus: " + event.Type)
	println("eventbus: " + event.Type)
}
