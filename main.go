package main

import (
	"bufio"
	"log"

	"github.com/jacobsa/go-serial/serial"
	"svenschwermer.de/ams-han-proxy/hdlc"
)

func main() {
	log.SetFlags(log.Lmicroseconds)

	options := serial.OpenOptions{
		PortName:        "/dev/ttyUSB0",
		BaudRate:        2400,
		DataBits:        8,
		ParityMode:      serial.PARITY_EVEN,
		StopBits:        1,
		MinimumReadSize: 1,
	}

	port, err := serial.Open(options)
	if err != nil {
		log.Fatal(err)
	}
	defer port.Close()

	scanner := bufio.NewScanner(port)
	scanner.Split(hdlc.SplitFrames)
	for scanner.Scan() {
		f := &hdlc.Frame{}
		if err := f.UnmarshalBinary(scanner.Bytes()); err != nil {
			log.Printf("Failed to decode frame: %s", err)
		} else {
			log.Print(f)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
