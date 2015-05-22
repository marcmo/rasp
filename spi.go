package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/platforms/gpio"
	"github.com/hybridgroup/gobot/platforms/raspi"
	"os"
	"time"
)

var channel = make(chan []byte)

type SpiState int

const (
	NO_CHIP_SELECT SpiState = 1 << iota
	CHIP_SELECT    SpiState = 2
	CLOCKING       SpiState = 4
)

func main() {
	virusZip := flag.String("z", "", "virus to kill AI")
	flag.Parse()
	if "" != *virusZip {
		fmt.Printf("uploading virusfile %s\n", *virusZip)
	}
	gbot := gobot.NewGobot()
	r := raspi.NewRaspiAdaptor("raspi")
	mosi := gpio.NewDirectPinDriver(r, "pin", "36")
	ss := gpio.NewDirectPinDriver(r, "pin", "37")
	sclk := gpio.NewDirectPinDriver(r, "pin", "38")
	cycleTime := 50 * time.Millisecond

	work := func() {

		spiState := NO_CHIP_SELECT
		var BitChannel = make(chan byte)
		go func() {
			for v := range channel {
				fmt.Printf("working on %v\n", v)
				for _, b := range v {
					for _, bit := range bitsInByte(b) {
						BitChannel <- bit
					}
				}
			}
		}()
		stream(*virusZip)
		// starting of with SS high
		ss.DigitalWrite(1)
		time.Sleep(10 * cycleTime)

		i := 0
		// start of with pulling clock down and writing data
		sclkSignal := byte(0)
		nextByte := byte(0)
		val := []byte{}
		data2send := false

		gobot.Every(cycleTime/2, func() {

			if !data2send && spiState == NO_CHIP_SELECT {
				select {
				case nextByte = <-BitChannel:
					data2send = true
					val = bitsInByte(nextByte)
				default:
					fmt.Printf("waiting for data\n")
				}
			}
			switch spiState {
			case NO_CHIP_SELECT:
				if data2send {
					// now activate the slave
					ss.DigitalWrite(0)
					spiState = CHIP_SELECT
				}
				break
			case CHIP_SELECT:
				if data2send {
					// start clocking
					// data only changes on the falling edge of SCLK and
					// is only read on the rising edge of SCLK
					sclk.DigitalWrite(1)
					spiState = CLOCKING
				} else {
					// now de-activate the slave
					ss.DigitalWrite(1)
					spiState = NO_CHIP_SELECT
				}
				break
			case CLOCKING:
				if data2send {
					sclk.DigitalWrite(sclkSignal)
					// only write on falling edge
					if 0 == sclkSignal {
						i = i + 1
						data2send = i < 8
						mosi.DigitalWrite(val[i])
					}
				} else {
					i = 0
					// only stop clocking after falling edge
					if 0 == sclkSignal {
						spiState = CHIP_SELECT
					}
				}
				break
			default:
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

func stream(f string) {
	fs, err := os.Open(f)
	check(err)
	defer fs.Close()
	r := bufio.NewReader(fs)
	for {
		done, content := streamFile(r, 10)
		channel <- content
		if done {
			break
		}
	}
}
func streamFile(r *bufio.Reader, chunkSize int) (bool, []byte) {
	s := make([]byte, chunkSize, chunkSize)
	c, err := r.Read(s)
	check(err)
	return (c < chunkSize), s
}
func check(e error) {
	if e != nil {
		panic(e)
	}
}
