package bus

type Message struct {
	Payload any
	replyTo chan Response
}

func (m Message) Reply(r Response) {
	if m.replyTo != nil {
		m.replyTo <- r
	}
}

type Response struct {
	Payload any
	Err     error
}
