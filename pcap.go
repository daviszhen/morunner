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

func analyzeDir(path string, filter, rexpr string) error {
	fInfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	regExpr, err := regexp.Compile(rexpr)
	if err != nil {
		return err
	}

	if !fInfo.IsDir() {
		return analyzePcap(path, filter, regExpr)
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
				err = analyzePcap(file, filter, regExpr)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func analyzePcap(fname string, filter string, regExpr *regexp.Regexp) error {

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
	for pkt := range pktSrc.Packets() {
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
		net := pkg.NetworkLayer().NetworkFlow()
		trans := pkg.TransportLayer().TransportFlow()
		if srcPort != "" && srcPort != trans.Src().String() {
			return nil
		}
		if dstPort != "" && dstPort != trans.Dst().String() {
			return nil
		}

		fmt.Println("net", net.Src(), ":", trans.Src(), "=>", net.Dst(), ":", trans.Dst())
		if displayBytesLimit > 0 {
			fmt.Println(string(cmdContent[:min(displayBytesLimit, cmdLen)]))
		}

	}
	return nil
}
