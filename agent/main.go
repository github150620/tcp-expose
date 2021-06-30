package main

import (
	"common/tcpjoin"
	"flag"
	"fmt"
	"log"
	"net"
	"time"
)

var (
	serverAddr    string
	proxyAddr     string
	proxyToken string
)

var (
	ch chan struct{}
)

func main() {
	flag.Usage = func() {
		fmt.Println("Usage: agent -server <address> -proxy <address> -token <token>")
		flag.PrintDefaults()
	}

	flag.StringVar(&serverAddr, "server", "", "The server address like 192.168.1.100:3389.")
	flag.StringVar(&proxyAddr, "proxy", "", "The proxy address like x.x.x.x:x.")
	flag.StringVar(&proxyToken, "token", "", "The token proxy will check.")
	flag.Parse()

	if serverAddr == "" {
		flag.Usage()
		return
	}

	if proxyAddr == "" {
		flag.Usage()
		return
	}

	if proxyToken == "" {
		flag.Usage()
		return
	}

	log.Println("agent v0.1.0")

	ch = make(chan struct{}, 5)
	ch <- struct{}{}
	ch <- struct{}{}
	ch <- struct{}{}
	ch <- struct{}{}
	ch <- struct{}{}

	for {
		select {
		case <-ch:
			go connectAndServe()
		}
	}
}

func connectAndServe() {
	defer func() {
		ch <- struct{}{}
	}()

	buf := make([]byte, len(proxyToken))

	rw1, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		time.Sleep(time.Second * 5)
		return
	}
	log.Printf("[INFO] %s connected to proxy", rw1.LocalAddr().String())

	_, err = rw1.Write([]byte(proxyToken))
	if err != nil {
		rw1.Close()
		time.Sleep(time.Second * 5)
		return
	}
	log.Printf("[INFO] %s sent password", rw1.LocalAddr().String())

	n, err := rw1.Read(buf)
	if err != nil {
		rw1.Close()
		time.Sleep(time.Second * 5)
		return
	}
	log.Printf("[INFO] %s received password %s", rw1.LocalAddr().String(), buf[:n])

	if string(buf[:n]) != proxyToken {
		rw1.Close()
		time.Sleep(time.Second * 5)
		return
	}

	rw2, err := net.Dial("tcp", serverAddr)
	if err != nil {
		rw1.Close()
		time.Sleep(time.Second * 5)
		return
	}
	log.Printf("[INFO] %s connected to server", rw1.LocalAddr().String())

	log.Printf("[INFO] %s <-> %s", rw1.LocalAddr().String(), rw2.LocalAddr().String())
	join := tcpjoin.New(rw1, rw2)
	go func() {
		join.Run()
		log.Printf("[INFO] %s -x- %s", rw1.LocalAddr().String(), rw2.LocalAddr().String())
	}()
}
