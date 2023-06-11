package leader

import "github.com/Ignaciocl/tp1SisdisCommons/client"

type Leader interface {
	WakeMeUpWhenSeptemberEnds() // Blocking method, will continue if the object is leader
	Run()                       // Will run forever, usable on go functions.
	Close()
}

func createSocketConfig(addr string) client.SocketConfig {
	return client.SocketConfig{
		Protocol:    "tcp",
		NodeAddress: addr,
		NodeACK:     "puto el que lee",
		PacketLimit: 8192,
	}
}
