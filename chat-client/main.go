package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {

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
		c.Close()
	}()

	go func() {
		_, err := io.Copy(c, os.Stdin)
		if err != nil {
			fmt.Println(err)
			if err == io.EOF {
				return
			}
		}
	}()
	go func() {
		_, err := io.Copy(os.Stdin, c)
		if err != nil {
			fmt.Println(err)
			if err == io.EOF {
				return
			}
		}
	}()

	<-sig
}
