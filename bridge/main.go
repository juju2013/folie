package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.bug.st/serial.v1"
)

var (
	mqttFlag  = flag.String("m", "tcp://127.0.0.1:1883", "MQTT broker")
	portFlag  = flag.String("p", "/dev/cu.usbmodem322F2211", "serial port")
	quietFlag = flag.Bool("q", false, "quiet mode, don't show in/out messages")
	topicFlag = flag.String("t", "bridge/%", "MQTT topic template")
)

func main() {
	flag.Parse()
	log.SetFlags(0) // omit timestamps

	dev, err := serial.Open(*portFlag, &serial.Mode{BaudRate: 115200})
	check(err)
	defer dev.Close()

	opts := mqtt.NewClientOptions().AddBroker(*mqttFlag)
	c := mqtt.NewClient(opts)
	t := c.Connect()
	mqttCheck(t)

	t = c.Subscribe(topic("out"), 0, func(c mqtt.Client, msg mqtt.Message) {
		if !*quietFlag {
			fmt.Printf("> %s\n", msg.Payload())
		}
		dev.Write(append(msg.Payload(), '\r'))
	})
	mqttCheck(t)

	go readSerial(dev, c)

	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		check(err)
		dev.Write([]byte(text[:len(text)-1] + "\r"))
	}
	// exit when stdin closes
}

func topic(s string) string {
	return strings.Replace(*topicFlag, "%", s, -1)
}

func readSerial(dev serial.Port, client mqtt.Client) {
	scanner := bufio.NewScanner(dev)
	for scanner.Scan() {
		msg := scanner.Text()
		if !*quietFlag {
			fmt.Printf("< %s\n", msg)
		}
		t := client.Publish(topic("in"), 0, false, msg)
		mqttCheck(t)
	}
	os.Exit(1)
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func mqttCheck(t mqtt.Token) {
	t.Wait()
	check(t.Error())
}
