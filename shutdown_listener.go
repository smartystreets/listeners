package listeners

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type ShutdownListener struct {
	mutex    sync.Once
	channel  chan os.Signal
	shutdown func()
}

func NewShutdownListener(shutdown func()) *ShutdownListener {
	channel := make(chan os.Signal, 16)
	signal.Notify(channel, os.Interrupt, syscall.SIGTERM)

	return &ShutdownListener{channel: channel, shutdown: shutdown}
}

func (this *ShutdownListener) Listen() {
	if message := <-this.channel; message != nil {
		log.Printf("[INFO] Received application shutdown signal [%s].\n", message)
	}

	this.shutdown()
}

func (this *ShutdownListener) Close() {
	this.mutex.Do(this.close)
}

func (this *ShutdownListener) close() {
	signal.Stop(this.channel)
	close(this.channel)
	log.Println("[INFO] Unsubscribed from further application shutdown signals.")
}
