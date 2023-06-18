package leader

type Leader interface {
	WakeMeUpWhenSeptemberEnds() // Blocking method, will continue if the object is leader
	Run()                       // Will run forever, usable on go functions.
	Close()
}
