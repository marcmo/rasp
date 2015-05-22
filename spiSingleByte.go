package main

import (
	"flag"
	"fmt"
	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/platforms/gpio"
	"github.com/hybridgroup/gobot/platforms/raspi"
	"time"
)

func oldmain() {
	byteToWrite := flag.Int("b", 100, "byte to send over spi")
	flag.Parse()
	gbot := gobot.NewGobot()
	r := raspi.NewRaspiAdaptor("raspi")
	mosi := gpio.NewDirectPinDriver(r, "pin", "36")
	ss := gpio.NewDirectPinDriver(r, "pin", "37")
	sclk := gpio.NewDirectPinDriver(r, "pin", "38")
	cycleTime := 100 * time.Millisecond

	work := func() {
		// starting of with SS high
		ss.DigitalWrite(1)
		fmt.Printf("SS HIGH\n")
		time.Sleep(1 * cycleTime)
		ss.DigitalWrite(0)
		time.Sleep(1 * cycleTime)
		ss.DigitalWrite(1)
		time.Sleep(1 * cycleTime)
		// start clocking
		sclk.DigitalWrite(0) // start LOW
		time.Sleep(1 * cycleTime)
		// now activate the slave
		ss.DigitalWrite(0)
		fmt.Printf("SS LOW\n")
		val := bitsInByte(byte(*byteToWrite))
		i := 0
		// start of with pulling clock down and writing data
		sclkSignal := byte(0)
		gobot.Every(cycleTime/2, func() {
			if i == 0 {
				// write BEFORE pulling up
				mosi.DigitalWrite(val[0])
				time.Sleep(1 * cycleTime)
				i = i + 1
			} else if i < 5*8 {
				if 0 == sclkSignal {
					sclkSignal = 1
					fmt.Printf("0->1\n")
					sclk.DigitalWrite(sclkSignal)
				} else {
					sclkSignal = 0
					fmt.Printf("1->0\n")
					sclk.DigitalWrite(sclkSignal)
					mosi.DigitalWrite(val[i%8])
					i = i + 1
				}
			} else {
				sclk.DigitalWrite(0)
				time.Sleep(1 * cycleTime)
				sclk.DigitalWrite(1)
				time.Sleep(1 * cycleTime)
				sclk.DigitalWrite(0)
				time.Sleep(1 * cycleTime)
				ss.DigitalWrite(1)
			}
		})
	}

	robot := gobot.NewRobot("spiBot", // robot name
		[]gobot.Connection{r},          // Connections which are automatically started and stopped with the robot
		[]gobot.Device{sclk, ss, mosi}, // Devices which are automatically started and stopped with the robot
		work) // work routine the robot will execute once all devices and connections have been started

	gbot.AddRobot(robot)
	gbot.Start()
}
