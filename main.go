package main

import (
	"net"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var opts struct {
	Port   string `short:"p" long:"port" default:":2203" description:"Source port to listen on"`
	Target string `short:"t" long:"target" description:"Target address to forward to" required:"true"`
	Buffer int    `short:"b" long:"buffer" default:"10240" description:"Buffer size for packets"`
	Dump   bool   `short:"d" long:"dump" description:"Dump packets to stdout"`
	Json   bool   `short:"j" long:"json" description:"Output JSON"`
	Hex    bool   `short:"x" long:"hex" description:"Output hex"`
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal().Err(err).Msg("Error parsing flags")
	}

	sourceAddr, err := net.ResolveUDPAddr("udp", opts.Port)
	if err != nil {
		log.Fatal().Err(err).Str("address", opts.Port).Msg("Error resolving source address")
	}

	targetAddr, err := net.ResolveUDPAddr("udp", opts.Target)
	if err != nil {
		log.Fatal().Err(err).Str("address", opts.Target).Msg("Error resolving target address")
	}

	listenConn, err := net.ListenUDP("udp", sourceAddr)
	if err != nil {
		log.Fatal().Err(err).Str("address", opts.Port).Msg("Error listening on address")
	}
	defer listenConn.Close()

	targetConn, err := net.DialUDP("udp", nil, targetAddr)
	if err != nil {
		log.Fatal().Err(err).Str("address", opts.Target).Msg("Error connecting to target")
	}
	defer targetConn.Close()

	log.Info().Str("source", sourceAddr.String()).Str("target", targetAddr.String()).Msgf("Listening on %s, forwarding to %s", sourceAddr, targetAddr)

	for {
		buffer := make([]byte, opts.Buffer)
		n, addr, err := listenConn.ReadFromUDP(buffer)

		if err != nil {
			log.Warn().Err(err).Msg("Error reading from UDP socket")
			continue
		}

		log.Info().Int("length", n).Str("source", addr.String()).Msg("Received packet")
		if opts.Dump {
			if opts.Json {
				log.Debug().RawJSON("data", buffer[:n]).Msg("Data (JSON)")
			} else if opts.Hex {
				log.Debug().Hex("data", buffer[:n]).Msg("Data (Hex)")
			} else {
				log.Debug().Str("data", strings.TrimSpace(string(buffer[:n]))).Msg("Data")
			}
		}

		_, err = targetConn.Write(buffer[:n])
		if err != nil {
			log.Warn().Err(err).Str("target", targetAddr.String()).Msg("Error writing to target")
		}
	}
}
