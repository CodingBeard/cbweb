package cbweb

type FlashMessage struct {
	Type    string
	Message string
}

type Flash struct {
	Messages map[string][]FlashMessage
}

func (f *Flash) AddMessage(group string, message FlashMessage) {
	if f.Messages == nil {
		f.Messages = make(map[string][]FlashMessage)
	}
	f.Messages[group] = append(f.Messages[group], message)
}

func (f *Flash) GetMessages(group string) []FlashMessage {
	if f.Messages == nil {
		f.Messages = make(map[string][]FlashMessage)
	}

	messages := f.Messages[group]

	delete(f.Messages, group)

	return messages
}

func (f *Flash) HasMessages(group string) bool {
	if f.Messages == nil {
		return false
	}

	return len(f.Messages[group]) > 0
}
