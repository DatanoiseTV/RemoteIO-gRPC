// Set up a connection to the server.
package main

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"os"
	"os/signal"
	remoteio "remoteio/rio"
	"time"
)

const (
	address     = "patchbox.local:9000"
	defaultName = "world"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := remoteio.NewRemoteIOClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.PinMode(ctx, &remoteio.PinModeMessage{Pin: 17, Mode: remoteio.PinModeMessage_DIGITAL_OUT})
	if err != nil {
		log.Fatal("Did not get message.")
	}
	log.Printf("Response: %v", r.GetPin())


	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	go func() {
		<-sigs
		fmt.Printf("You pressed ctrl + C. User interrupted infinite loop.")
		os.Exit(0)
	}()
	BlinkLED(c)



}

func BlinkLED(c remoteio.RemoteIOClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	state := false
	for {
		r, err := c.DigitalWrite(ctx, &remoteio.DigitalState{Pin: 17, State: state})
		if err != nil {
			log.Println("Did not get response.")
		} else {
			log.Printf("Response: %v", r.GetPin())
		}

		state = !state
		time.Sleep(time.Millisecond * 50)
	}
}