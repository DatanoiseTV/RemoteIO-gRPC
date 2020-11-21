// Set up a connection to the server.
package main

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	remoteio "remoteio/rio"
	"time"
)

const (
	address     = "localhost:9000"
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

	r, err := c.PinMode(ctx, &remoteio.PinModeMessage{Pin: 1, Mode: remoteio.PinModeMessage_ANALOG_IN})
	if err != nil {
		log.Fatal("Did not get message.")
	}
	log.Printf("Response: %v", r.GetPin())

}