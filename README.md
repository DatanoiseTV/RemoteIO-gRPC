# RemoteIO-gRPC
Remote I/O service for RaspberryPi using gRPC (and Protobuf).
Allows to use SPI, Analog I/O (with external ADC/DAC) and Digital I/O from any client and language supporting gRPC.

## Installation
```
go get github.com/DatanoiseTV/RemoteIO-gRPC
```

## Why?

Sometimes you are working on a project or need to test a new electronics component and need some GPIO quickly
usable from your main computer. Running this software gives you full access to the GPIO, I2C and SPI peripherals
on the Pi3/4.

## Internals
It is based on Googles gRPC framework, which they use to power most of their inter-server/inter-service communication.
gRPC provides a RPC (Remote Procedure Call) interface, which allows to call remote functions / procedures. The serialization
magic behind that is Google Protobuf.

## Protocol Implementation
```
service RemoteIO {
  rpc pinMode(PinModeMessage) returns (PinModeMessage){};
  rpc digitalRead(DigitalState) returns (DigitalState){};
  rpc digitalWrite(DigitalState) returns (DigitalState){};
  rpc analogRead(AnalogState) returns (AnalogState){};
  rpc analogWrite(AnalogState) returns (AnalogState){};

  rpc spiRead(SPIMessage) returns (SPIMessage){};
  rpc spiWrite(SPIMessage) returns (SPIMessage){};

  rpc subscribeInterrupt(InterruptMessage) returns (stream DigitalState){};
}
```

## Notes
If you are into very time-critical stuff, don't use WiFi. This should work at a few (hundred) kHz I/O rate (analog+digital) via ethernet. Exact numbers to be determined.
It sets the process priority to -20, to get the best performance on a PREEMPT_RT kernel.
