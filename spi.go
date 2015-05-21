package main

import (
	"flag"
	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/platforms/gpio"
	"github.com/hybridgroup/gobot/platforms/raspi"
	"time"
)

func main() {
	byteToWrite := flag.Int("b", 200, "byte to send over spi")
	gbot := gobot.NewGobot()
	r := raspi.NewRaspiAdaptor("raspi")
	mosi := gpio.NewDirectPinDriver(r, "pin", "36")
	ss := gpio.NewDirectPinDriver(r, "pin", "37")
	sclk := gpio.NewDirectPinDriver(r, "pin", "38")
	cycleTime := 50 * time.Millisecond

	work := func() {
		// starting of with SS high
		ss.DigitalWrite(1)
		time.Sleep(10 * cycleTime)
		// now activate the slave
		ss.DigitalWrite(0)
		time.Sleep(cycleTime)
		// start clocking
		// data only changes on the falling edge of SCLK and
		// is only read on the rising edge of SCLK
		sclk.DigitalWrite(1)
		time.Sleep(cycleTime)
		val := bitsInByte(byte(*byteToWrite))
		i := 0
		// start of with pulling clock down and writing data
		sclkSignal := byte(0)
		gobot.Every(cycleTime/2, func() {
			if i < 10*8 {
				i = i + 1
				sclk.DigitalWrite(sclkSignal)
				if 0 == sclkSignal {
					// only write on falling edge
					mosi.DigitalWrite(val[i%8])
				}
			}
			sclkSignal = toggle(sclkSignal)
		})
	}

	robot := gobot.NewRobot("spiBot", // robot name
		[]gobot.Connection{r},          // Connections which are automatically started and stopped with the robot
		[]gobot.Device{sclk, ss, mosi}, // Devices which are automatically started and stopped with the robot
		work) // work routine the robot will execute once all devices and connections have been started

	gbot.AddRobot(robot)
	gbot.Start()
}

func toggle(i byte) byte {
	if i == 0 {
		return 1
	} else {
		return 0
	}
}
