package engine

type MessageHandler interface {
	Send(assertion *Assertion)
}

func StartListener(handler MessageHandler, c chan *Assertion, quit chan bool) {
	// fmt.Println("start listener")
	for {
		select {
		case assertion := <-c:
			// fmt.Println("sending....")
			handler.Send(assertion)
		case <-quit:
			// fmt.Println("end listener")
			return
		}
	}

}
