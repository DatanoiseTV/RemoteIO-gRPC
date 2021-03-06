package main

import (
	"context"
	"fmt"
	remoteio "github.com/DatanoiseTV/RemoteIO-gRPC-proto"
	"github.com/d2r2/go-i2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	//log.Printf("Pin mode: Pin %v: %v", in.GetPin(), in.GetMode())

	pin := rpio.Pin(in.GetPin())
	if in.GetMode() == remoteio.PinModeMessage_ANALOG_IN || in.GetMode() == remoteio.PinModeMessage_DIGITAL_IN {
		pin.Input()
	} else if in.GetMode() == remoteio.PinModeMessage_ANALOG_OUT || in.GetMode() == remoteio.PinModeMessage_DIGITAL_OUT{
		pin.Output()
	}

	now := timestamppb.Now()
	return &remoteio.PinModeMessage{Pin: in.GetPin(), Mode: in.GetMode(), Timestamp: now}, nil
}

func (s *server) DigitalRead(ctx context.Context, in *remoteio.DigitalState) (*remoteio.DigitalState, error){
	//log.Printf("DigitalRead: Pin %v", in.GetPin())
	pin := rpio.Pin(in.GetPin())
	state := false
	if pin.Read() == rpio.Low {
		state = false
	} else if pin.Read() == rpio.High {
		state = true
	}
	now := timestamppb.Now()
	return &remoteio.DigitalState{Pin: in.GetPin(), State: state, Timestamp: now}, nil
}

func (s *server) DigitalWrite(ctx context.Context, in *remoteio.DigitalState) (*remoteio.DigitalState, error){
	//log.Printf("DigitalWrite: Pin %v: %v", in.GetPin(), in.GetState())
	pin := rpio.Pin(in.GetPin())
	if in.GetState() == false {
		pin.Write(rpio.Low)
	} else if in.GetState() == true {
		pin.Write(rpio.High)
	}
	now := timestamppb.Now()
	return &remoteio.DigitalState{Pin: in.GetPin(), State: in.GetState(), Timestamp: now}, nil
}

func (s *server) AnalogRead(ctx context.Context, in *remoteio.AnalogState) (*remoteio.AnalogState, error){
	//log.Printf("AnalogRead: Pin %v", in.GetPin())
	now := timestamppb.Now()
	return &remoteio.AnalogState{Pin: in.GetPin(), Value: 0, Timestamp: now}, nil
}

func (s *server) AnalogWrite(ctx context.Context, in *remoteio.AnalogState) (*remoteio.AnalogState, error){
	//log.Printf("AnalogWrite: Pin %v: %v", in.GetPin(), in.GetValue())
	pin := rpio.Pin(in.GetPin())
	pin.Mode(rpio.Pwm)
	pin.Freq(64000)
	pin.DutyCycle(in.GetValue() & 0xFF, 256)
	now := timestamppb.Now()
	return &remoteio.AnalogState{Pin: in.GetPin(), Value: 0, Timestamp: now}, nil
}

func (s *server) SpiRead(ctx context.Context, in *remoteio.SPIMessage) (*remoteio.SPIMessage, error){
	buffer := in.GetBytes()
	buffer_u8 := make([]byte, len(buffer))

	for i := 0; i<len(buffer)-1; i++{
		buffer_u8[i] = byte(buffer[i])
	}
	//log.Printf("SPIRead: %v", in.GetBytes())
	if err := rpio.SpiBegin(rpio.Spi0); err != nil {
		log.Println("Could not use SPI.")
	}
	if in.GetCs() >= 0 { rpio.SpiChipSelect(uint8(in.GetCs())) }
	if in.GetSpeed() >= 1 && in.GetSpeed() <= 18000000 { rpio.SpiSpeed(int(in.GetSpeed())) }

	rpio.SpiExchange(buffer_u8);

	rpio.SpiEnd(rpio.Spi0)

	buffer = make([]uint32, len(buffer_u8))
	for i := 0; i<len(buffer_u8)-1; i++ {
		buffer[i] = uint32(buffer_u8[i])
	}
	now := timestamppb.Now()
	return &remoteio.SPIMessage{Bytes: buffer, Timestamp: now}, nil
}
func (s *server) I2CRead(ctx context.Context, in *remoteio.I2CMessage) (*remoteio.I2CMessage, error){
	i2c, err := i2c.NewI2C(byte(in.GetAddr()), 1)
	if err != nil { log.Println(err) }
	defer i2c.Close()

	buffer := in.GetBytes()
	buffer_u8 := make([]byte, len(buffer))

	for i := 0; i<len(buffer)-1; i++{
		buffer_u8[i] = byte(buffer[i])
	}

	_, err = i2c.WriteBytes(buffer_u8)
	if err != nil { log.Println(err) }

	ret := []byte{}
	i2c.ReadBytes(ret)

	if len(ret) > 0{
		buffer = make([]uint32, len(ret))
		for i := 0; i<len(buffer_u8)-1; i++ {
			buffer[i] = uint32(ret[i])
		}
	} else {
		buffer = []uint32{0x0}
	}

	now := timestamppb.Now()
	return &remoteio.I2CMessage{Bytes: buffer, Timestamp: now}, nil
}


func (s *server) SubscribeInterrupt(in *remoteio.InterruptMessage, src remoteio.RemoteIO_SubscribeInterruptServer) error {
	pin := rpio.Pin(in.GetPin())
	pin.Input()
	pin.PullUp()

	switch in.GetTriggerType(){
	case remoteio.InterruptMessage_RISING_EDGE:
		pin.Detect(rpio.RiseEdge)
		break
	case remoteio.InterruptMessage_FALLING_EDGE:
		pin.Detect(rpio.FallEdge)
		break
	case remoteio.InterruptMessage_BOTH_EDGE:
		pin.Detect(rpio.AnyEdge)
		break
	}

	for {
		if pin.EdgeDetected() {
			pinState := pin.ReadPull()
			boolState := false
			if(pinState == 0){ boolState = false }
			if(pinState == 1){ boolState = true }

			src.Send(&remoteio.DigitalState{Pin: in.GetPin(), State: boolState, Timestamp: timestamppb.Now()})
		}
	}
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
	reflection.Register(grpcServer)

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
