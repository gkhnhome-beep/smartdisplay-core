package audit

type Entry struct {
	Timestamp string
	Action    string
	Detail    string
}

var entries []Entry

func Record(action, detail string) {
	entry := Entry{
		Timestamp: now(),
		Action:    action,
		Detail:    detail,
	}
	entries = append(entries, entry)
}

func GetEntries() []Entry {
	return append([]Entry(nil), entries...)
}

func now() string {
	// Use time.Now().Format for ISO8601
	return "2006-01-02T15:04:05Z" // placeholder, replace with time.Now().Format(time.RFC3339) in real code
}
