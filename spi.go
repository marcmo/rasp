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

type SpiState int

const (
	NO_CHIP_SELECT    SpiState = 1 << iota
	CHIP_SELECT       SpiState = 2
	FIRST_DATA_ON_BUS SpiState = 4
	CLOCKING          SpiState = 8
)

func main() {
	virusZip := flag.String("z", "", "virus to kill AI")
	byteToWrite := flag.Int("b", 0xFFFF, "sample byte to send over spi")
	flag.Parse()
	var channel = make(chan []byte)
	if "" != *virusZip {
		fmt.Printf("uploading virusfile %s\n", *virusZip)
		go stream(*virusZip, channel)
	} else if *byteToWrite <= 0xFF {
		go func() {
			channel <- []byte{byte(*byteToWrite)}
		}()
	} else {
		fmt.Printf("usage: -z <zipfile.zip>\n")
		return
	}
	gbot := gobot.NewGobot()
	r := raspi.NewRaspiAdaptor("raspi")
	mosi := gpio.NewDirectPinDriver(r, "pin", "36")
	ss := gpio.NewDirectPinDriver(r, "pin", "37")
	sclk := gpio.NewDirectPinDriver(r, "pin", "38")
	selectSlave := func() { ss.DigitalWrite(0) }
	deselectSlave := func() { ss.DigitalWrite(1) }

	cycleTime := 100 * time.Millisecond

	work := func() {

		spiState := NO_CHIP_SELECT
		var ByteChannel = make(chan byte)
		go func() {
			for v := range channel {
				fmt.Printf("working on %v\n", v)
				for _, b := range v {
					ByteChannel <- b
				}
			}
		}()
		// starting of with SS high
		deselectSlave()
		time.Sleep(10 * cycleTime)

		i := 0
		// start of with pulling clock down and writing data
		nextClockSignal := byte(0)
		nextByte := byte(0)
		val := []byte{}
		data2send := false

		gobot.Every(cycleTime/2, func() {
			// wait until we are idle before pulling new data
			if !data2send && spiState == NO_CHIP_SELECT {
				select {
				case nextByte = <-ByteChannel:
					data2send = true
					val = bitsInByte(nextByte)
				default:
					fmt.Printf("waiting for data\n")
				}
			}
			switch spiState {
			case NO_CHIP_SELECT:
				if data2send {
					selectSlave()
					spiState = CHIP_SELECT
				}
				break
			case CHIP_SELECT:
				if data2send {
					// write BEFORE pulling up
					mosi.DigitalWrite(val[0])
					spiState = FIRST_DATA_ON_BUS
				} else {
					deselectSlave()
					spiState = NO_CHIP_SELECT
				}
				break
			case FIRST_DATA_ON_BUS:
				if data2send {
					// start clocking
					// by indicating SCLK UP
					nextClockSignal = 1
					sclk.DigitalWrite(nextClockSignal)
					spiState = CLOCKING
				}
				break
			case CLOCKING:
				if data2send {
					sclk.DigitalWrite(nextClockSignal)
					// only write on falling edge
					if 0 == nextClockSignal {
						mosi.DigitalWrite(val[i])
						i = i + 1
						data2send = i < 8
					}
				} else {
					i = 0
					// only stop clocking after falling edge
					if 1 == nextClockSignal {
						spiState = CHIP_SELECT
					} else {
						sclk.DigitalWrite(nextClockSignal)
					}
				}
				break
			default:
			}
			nextClockSignal = toggle(nextClockSignal)
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

func stream(f string, c chan<- []byte) {
	fs, err := os.Open(f)
	check(err)
	defer fs.Close()
	r := bufio.NewReader(fs)
	for {
		done, content := streamFile(r, 10)
		c <- content
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
