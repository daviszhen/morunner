package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"go.uber.org/zap"
)

func analyzeDir(sigs chan os.Signal, path string, filter, rexpr string) error {
	fInfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	regExpr, err := regexp.Compile(rexpr)
	if err != nil {
		return err
	}

	if !fInfo.IsDir() {
		return analyzePcap(sigs, path, filter, regExpr)
	} else {
		files := make([]string, 0)
		err = filepath.Walk(path, func(x string, info os.FileInfo, err error) error {
			files = append(files, x)
			return nil
		})
		if err != nil {
			return err
		}
		for _, file := range files {
			if strings.Contains(file, ".pcap") {
				err = analyzePcap(sigs, file, filter, regExpr)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func analyzePcap(sigs chan os.Signal, fname string, filter string, regExpr *regexp.Regexp) error {

	fmt.Println("analyze... ", fname)
	reader, err := pcap.OpenOffline(fname)
	if err != nil {
		logger.Error("PCAP OpenOffline error (handle to read packet):", zap.String("fname", fname), zap.Error(err))
		return err
	}
	defer reader.Close()

	err = reader.SetBPFFilter(filter)
	if err != nil {
		return err
	}

	pktSrc := gopacket.NewPacketSource(reader, reader.LinkType())
	quit := false
	for pkt := range pktSrc.Packets() {
		select {
		case <-sigs:
			quit = true
		default:
		}
		if quit {
			break
		}
		tcpLayer := pkt.Layer(layers.LayerTypeTCP)
		tcpPayload := tcpLayer.LayerPayload()
		err = handleMysql(pkt, tcpPayload, regExpr)
		if err != nil {
			return err
		}
	}
	return err
}

const (
	mysqlPacketFixLen = 5 //header len (4 bytes) + cmd (1 byte)
)

// find load data local
func handleMysql(pkg gopacket.Packet, payload []byte, regexpr *regexp.Regexp) error {
	if len(payload) <= mysqlPacketFixLen {
		return nil
	}
	cmdContent := payload[mysqlPacketFixLen:]
	cmdLen := len(cmdContent)
	if regexpr.Match(cmdContent) {
		ts := pkg.Metadata().Timestamp
		net := pkg.NetworkLayer().NetworkFlow()
		trans := pkg.TransportLayer().TransportFlow()

		portFilter := false
		if srcPort == "" && dstPort == "" {

		} else if srcPort == "" && dstPort != "" {
			portFilter = dstPort != trans.Dst().String()
		} else if srcPort != "" && dstPort == "" {
			portFilter = srcPort != trans.Src().String()
		} else {
			portFilter = srcPort != trans.Src().String() || dstPort != trans.Dst().String()
		}

		if portFilter {
			return nil
		}

		fmt.Println()
		fmt.Println("cap ts", ts, "net", net.Src(), ":", trans.Src(), "=>", net.Dst(), ":", trans.Dst())
		if displayBytesLimit > 0 {
			dlen := min(displayBytesLimit, cmdLen)
			if displayBinary {
				fmt.Print("bin:")
				for _, x := range cmdContent[:dlen] {
					fmt.Printf("%x ", x)
				}
				fmt.Println()
			}
			if displayText {
				fmt.Println("text:", string(cmdContent[:dlen]))
			}
		}
	}
	return nil
}
