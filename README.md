# RemoteIO-gRPC
Remote I/O service for RaspberryPi using gRPC (and Protobuf).
Allows to use SPI, Analog I/O (with external ADC/DAC) and Digital I/O from any client and language supporting gRPC.

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
}
```

## Notes
If you are into very time-critical stuff, don't use WiFi. This should work at a few (hundred) kHz I/O rate (analog+digital) via ethernet. Exact numbers to be determined.
It sets the process priority to -20, to get the best performance on a PREEMPT_RT kernel.
