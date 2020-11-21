package main

import (
	"context"
	"fmt"
	remoteio "github.com/DatanoiseTV/RemoteIO-gRPC-proto"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/stianeikeland/go-rpio"
)

type server struct {
	remoteio.UnimplementedRemoteIOServer
}

func (s *server) PinMode(ctx context.Context, in *remoteio.PinModeMessage) (*remoteio.PinModeMessage, error){
	log.Printf("Pin mode: %v %v", in.GetPin(), in.GetMode())

	pin := rpio.Pin(in.GetPin())
	if in.GetMode() == remoteio.PinModeMessage_ANALOG_IN || in.GetMode() == remoteio.PinModeMessage_DIGITAL_IN {
		pin.Input()
	} else if in.GetMode() == remoteio.PinModeMessage_ANALOG_OUT || in.GetMode() == remoteio.PinModeMessage_DIGITAL_OUT{
		pin.Output()
	}

	return &remoteio.PinModeMessage{Pin: in.GetPin(), Mode: in.GetMode()}, nil
}

func (s *server) DigitalRead(ctx context.Context, in *remoteio.DigitalState) (*remoteio.DigitalState, error){
	log.Printf("DigitalRead: %v", in.GetPin())
	pin := rpio.Pin(in.GetPin())
	state := false
	if pin.Read() == rpio.Low {
		state = false
	} else if pin.Read() == rpio.High {
		state = true
	}
	return &remoteio.DigitalState{Pin: in.GetPin(), State: state}, nil
}

func (s *server) DigitalWrite(ctx context.Context, in *remoteio.DigitalState) (*remoteio.DigitalState, error){
	log.Printf("DigitalWrite: %v, %v", in.GetPin(), in.GetState())
	pin := rpio.Pin(in.GetPin())
	if in.GetState() == false {
		pin.Write(rpio.Low)
	} else if in.GetState() == true {
		pin.Write(rpio.High)
	}
	return &remoteio.DigitalState{Pin: in.GetPin(), State: in.GetState()}, nil
}

func (s *server) AnalogRead(ctx context.Context, in *remoteio.AnalogState) (*remoteio.AnalogState, error){
	log.Printf("AnalogRead: %v", in.GetPin())
	return &remoteio.AnalogState{Pin: in.GetPin(), Value: 0}, nil
}

func (s *server) AnalogWrite(ctx context.Context, in *remoteio.AnalogState) (*remoteio.AnalogState, error){
	log.Printf("AnalogWrite: %v", in.GetPin())
	pin := rpio.Pin(in.GetPin())
	pin.Mode(rpio.Pwm)
	pin.Freq(64000)
	pin.DutyCycle(in.GetValue() & 0xFF, 256)
	return &remoteio.AnalogState{Pin: in.GetPin(), Value: 0}, nil
}

func (s *server) SPIRead(ctx context.Context, in *remoteio.SPIMessage) (*remoteio.SPIMessage, error){
	buffer := in.GetBytes()
	buffer_u8 := []byte{}
	for i := 0; i<len(buffer); i++{
		buffer_u8[i] = byte(buffer[i])
	}
	log.Printf("SPIRead: %v", in.GetBytes())
	if err := rpio.SpiBegin(rpio.Spi0); err != nil {
		log.Println("Could not use SPI.")
	}
	if in.GetCs() >= 0 { rpio.SpiChipSelect(uint8(in.GetCs())) }
	if in.GetSpeed() >= 1000000 && in.GetSpeed() <= 16000000 { rpio.SpiSpeed(int(in.GetSpeed())) }

	rpio.SpiExchange(buffer_u8);
	buffer = []uint32{}
	for i := 0; i<len(buffer_u8); i++ {
		buffer[i] = uint32(buffer_u8[i])
	}
	return &remoteio.SPIMessage{Bytes: buffer}, nil
}

func main() {
	if os.Getuid() != 0 {
		fmt.Println("Sorry, root required.")
		os.Exit(1)
	}

	syscall.Setpgid(0, 0); syscall.Setpriority(syscall.PRIO_PGRP, 0, -20)

	if err := rpio.Open(); err != nil {
		panic(err)
	}

	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	remoteio.RegisterRemoteIOServer(grpcServer, &server{})
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	go func() {
		<-sigs
		fmt.Printf("You pressed ctrl + C. User interrupted infinite loop.")
		os.Exit(0)
	}()


}
