// Set up a connection to the server.
package main

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"math/rand"

	//"math/rand"
	"os"
	"os/signal"
	remoteio "github.com/DatanoiseTV/RemoteIO-gRPC-proto"
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

	r, err := c.PinMode(ctx, &remoteio.PinModeMessage{Pin: 12, Mode: remoteio.PinModeMessage_ANALOG_OUT})
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

	fpsCounter := 0

	BlinkLED(c, fpsCounter)



}

func BlinkLED(c remoteio.RemoteIOClient, fps int) {

	state := false

	latency_avg := int64(0)

	go func() {
		ticker := time.NewTicker(time.Second)

		for{
			select {
			case <-ticker.C:
				log.Printf("FPS: %v", fps)
				log.Printf("Average latency (ms): %v", float32(latency_avg)/1000.0)
				fps = 0
				latency_avg = 0

			}
		}
	}();

	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		now := timestamppb.Now()
		r, err := c.AnalogWrite(ctx, &remoteio.AnalogState{Pin: 12, Value: uint32(rand.Intn(254)), Timestamp: now })
		//r, err := c.DigitalWrite(ctx, &remoteio.DigitalState{Pin: 12, State: state, Timestamp: now})
		if err != nil {
			log.Println("Did not get response.")
		} else {
			receiveTime := r.GetTimestamp().AsTime()
			latency := receiveTime.Sub(now.AsTime()).Milliseconds()
			if 1 == 0 {
				log.Printf("Latency: %v", latency)

			}
			latency_avg += latency

		}

		fps++

		state = !state
		//time.Sleep(time.Millisecond * 100)
	}
}