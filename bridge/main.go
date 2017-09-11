package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.bug.st/serial.v1"
)

func main() {
	log.SetFlags(0) // omit timestamps

	port := "/dev/cu.usbmodem322F2211"
	mode := serial.Mode{BaudRate: 115200}

	dev, err := serial.Open(port, &mode)
	check(err)

	defer dev.Close()

	go SerialReader(dev)
	dev.Write([]byte("1 2 + .\r"))

	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://127.0.0.1:1883")
	opts.SetDefaultPublishHandler(pubHandler)

	c := mqtt.NewClient(opts)
	if t := c.Connect(); t.Wait() && t.Error() != nil {
		panic(t.Error())
	}

	t := c.Subscribe("test/bridge", 0, nil)
	mqttCheck(t)

	for i := 0; i < 5; i++ {
		text := fmt.Sprintf("this is msg #%d!", i)
		t := c.Publish("test/bridge", 0, false, text)
		mqttCheck(t)
	}

	time.Sleep(1 * time.Second)
	log.Print("Done")
}

func pubHandler(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("MQTT: %s - %s\n", msg.Topic(), msg.Payload())
}

func SerialReader(dev serial.Port) {
	scanner := bufio.NewScanner(dev)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
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
