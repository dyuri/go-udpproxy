package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Port   string `short:"p" long:"port" default:":2203" description:"Source port to listen on"`
	Target string `short:"t" long:"target" description:"Target address to forward to" required:"true"`
	Buffer int    `short:"b" long:"buffer" default:"10240" description:"Buffer size for packets"`
	Dump   bool   `short:"d" long:"dump" description:"Dump packets to stdout"`
	Json   bool   `short:"j" long:"json" description:"Output JSON"`
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Printf("Error parsing flags: %s", err)
		os.Exit(1)
	}

	sourceAddr, err := net.ResolveUDPAddr("udp", opts.Port)
	if err != nil {
		log.Fatalf("Error resolving source address: %s", err)
	}

	targetAddr, err := net.ResolveUDPAddr("udp", opts.Target)
	if err != nil {
		log.Fatalf("Error resolving target address: %s", err)
	}

	listenConn, err := net.ListenUDP("udp", sourceAddr)
	if err != nil {
		log.Fatalf("Error listening on %s: %s", sourceAddr, err)
	}
	defer listenConn.Close()

	targetConn, err := net.DialUDP("udp", nil, targetAddr)
	if err != nil {
		log.Fatalf("Error connecting to %s: %s", targetAddr, err)
	}
	defer targetConn.Close()

	log.Printf("Listening on %s, forwarding to %s", sourceAddr, targetAddr)

	for {
		buffer := make([]byte, opts.Buffer)
		n, addr, err := listenConn.ReadFromUDP(buffer)

		if err != nil {
			log.Printf("Error reading from %s: %s", addr, err)
			continue
		}

		log.Printf("Received %d bytes from %s", n, addr)
		if opts.Dump {
			if opts.Json {
				var msg bytes.Buffer
				err = json.Indent(&msg, buffer[:n], "", "  ")
				if err != nil {
					log.Printf("Wrong JSON?\n%s", buffer[:n])
				} else {
					log.Printf("\n%s", msg.String())
				}
			} else {
				log.Printf("Data: %s", strings.TrimSpace(string(buffer[:n])))
			}
		}

		_, err = targetConn.Write(buffer[:n])
		if err != nil {
			log.Printf("Error writing to %s: %s", targetAddr, err)
		}
	}
}
