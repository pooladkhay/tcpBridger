package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gordonklaus/portaudio"
)

func main() {
	portaudio.Initialize()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("please server address as host:port")
		return
	}

	addr, err := net.ResolveTCPAddr("tcp", arguments[1])
	if err != nil {
		log.Fatalln(err)
	}

	c, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Fatalln(err)
	}

	defer func() {
		fmt.Println("\nexiting...")
		portaudio.Terminate()
		c.Close()
	}()

	go func() {
		for {
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			c.Write([]byte(text))
			if strings.Contains(text, "START_PARTY") {
				fmt.Println("breaking 1")
				break
			}
		}
	}()

	go func() {
		buf := make([]byte, 2000)
		for {
			io.ReadAtLeast(c, buf, 2)
			fmt.Println(string(buf))
			if strings.Contains(string(buf), "init") {
				fmt.Println("breaking 2")
				break
			}
			buf = make([]byte, 2000)
		}
		go sender(c, sig)
		go receiver(c, sig)
	}()

	<-sig
}

func receiver(c *net.TCPConn, sig <-chan os.Signal) {

	h, err := portaudio.DefaultHostApi()
	errCheck(err)
	p := portaudio.HighLatencyParameters(nil, h.DefaultOutputDevice)
	p.Output.Channels = 1

	stream, err := portaudio.OpenStream(p, func(in, out []float32) {
		fmt.Println("rx")
		errCheck(binary.Read(c, binary.BigEndian, out))
	})
	errCheck(err)

	errCheck(stream.Start())

	<-sig
	stream.Close()
	fmt.Println("\nExiting...")
	errCheck(stream.Stop())
}

func sender(c *net.TCPConn, sig <-chan os.Signal) {
	h, err := portaudio.DefaultHostApi()
	errCheck(err)
	p := portaudio.HighLatencyParameters(h.DefaultInputDevice, nil)
	p.Input.Channels = 1

	stream, err := portaudio.OpenStream(p, func(in, out []float32) {
		fmt.Println("ssx")
		errCheck(binary.Write(c, binary.BigEndian, in))
	})
	errCheck(err)

	errCheck(stream.Start())

	<-sig
	stream.Close()
	fmt.Println("\nExiting...")
	errCheck(stream.Stop())
}

func errCheck(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
