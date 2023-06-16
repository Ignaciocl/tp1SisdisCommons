package main

import (
	"fmt"
	"github.com/Ignaciocl/tp1SisdisCommons/leader"
	"os"
	"time"
)

func main() {
	id := os.Getenv("ID")
	name := fmt.Sprintf("maybe%s:3000", id)
	m := map[string]string{
		"1": "main:3000",
		"2": "maybe2:3000",
		"3": "maybe3:3000",
		"4": "maybe4:3000",
	}
	b, _ := leader.NewBully(id, name, "tcp", m)
	go func() { b.Run() }()
	for {
		b.WakeMeUpWhenSeptemberEnds()
		fmt.Printf("i am leader %s\n", id)
		time.Sleep(3 * time.Second)
	}
}
