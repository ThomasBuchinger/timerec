package api

type Event struct {
	Name    string
	Message string
	Target  string
	User    string
}

type Message struct {
	Message string
	User    string
}
