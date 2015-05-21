package main

import (
	"fmt"
	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/platforms/gpio"
	"github.com/hybridgroup/gobot/platforms/raspi"
	"time"
)

func main() {
	gbot := gobot.NewGobot()
	r := raspi.NewRaspiAdaptor("raspi")
	pin := gpio.NewDirectPinDriver(r, "pin", "40")
	work := func() {
		level := byte(1)

		gobot.Every(100*time.Millisecond, func() {
			fmt.Printf("ping...%d\n", level)
			pin.DigitalWrite(level)
			if level == 1 {
				level = 0
			} else {
				level = 1
			}
		})
	}

	robot := gobot.NewRobot("pinBot", []gobot.Connection{r}, []gobot.Device{pin}, work)

	gbot.AddRobot(robot)
	fmt.Printf("starting...\n")
	gbot.Start()
	fmt.Printf("started...\n")
}
