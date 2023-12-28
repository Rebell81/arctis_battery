package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/getlantern/systray"
	"github.com/sstallion/go-hid"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	selectedHidDevice *hid.Device
	errorNotConnected = errors.New("headset not connected")
	wasDisconnected   = false
)

func main() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   "log.log",
		MaxSize:    5, // megabytes
		MaxBackups: 3,
		MaxAge:     14,    //days
		Compress:   false, // disabled by default
	})

	initHideDevice()

	systray.Run(onReady, nil)
}

func onReady() {
	mQuit := systray.AddMenuItem("Quit", "Quit app")

	go func() {
		<-mQuit.ClickedCh
		_ = selectedHidDevice.Close()

		os.Exit(0)
	}()

	go loop()
}

func loop() {
	for {
		time.Sleep(time.Duration(5) * time.Second)

		if wasDisconnected {
			reInit()
			time.Sleep(time.Duration(1) * time.Second)
			setBat()

			continue
		}
		setBat()
	}
}

func setBat() {
	batValue, err := getBattery(selectedHidDevice)
	if err != nil {
		if errors.Is(err, errorNotConnected) {
			batValue = 0
			wasDisconnected = true
			log.Println(fmt.Sprintf("bat: %d	| err: %s", batValue, err))
		} else {
			log.Println(fmt.Sprintf("bat: %d	| err: %s", batValue, err))
		}
	} else {
		wasDisconnected = false
		log.Println(fmt.Sprintf("bat: %d", batValue))
	}

	var bytes = read(fmt.Sprintf("Icons/%d.ico", batValue))
	systray.SetIcon(bytes)
}

func read(fileName string) []byte {
	file, err := os.Open(fileName)
	if err != nil {
		log.Println(fmt.Errorf("err: %w", err))

		return nil
	}
	defer file.Close()

	// Get the file size
	stat, err := file.Stat()
	if err != nil {
		log.Println(fmt.Errorf("err: %w", err))

		return nil
	}

	// Read the file into a byte slice
	bs := make([]byte, stat.Size())
	_, err = bufio.NewReader(file).Read(bs)
	if err != nil && err != io.EOF {
		log.Println(fmt.Errorf("err: %w", err))

		return nil
	}

	return bs
}

func getBattery(d *hid.Device) (int, error) {
	_, err := d.Write([]byte{0x06, 0x14})
	if err != nil {
		return 0, err
	}
	report := make([]byte, 31)
	_, err = d.Read(report)
	if err != nil {
		return 0, err
	}
	if report[2] != 0x03 {
		return 0, errorNotConnected
	}

	_, err = d.Write([]byte{0x06, 0x18})
	if err != nil {
		return 0, err
	}
	_, err = d.Read(report)
	if err != nil {
		return 0, err
	}

	return int(report[2]), nil
}

func initHideDevice() {
	if err := hid.Init(); err != nil {
		log.Fatalf("Error when initializing HID library: %v", err)
	}
	steelHids := make([]*hid.DeviceInfo, 0)

	//[1038, 0x12ad], // Arctis 7 2019
	//[1038, 0x1260], // Arctis 7 2017
	//[1038, 0x1252], // Arctis Pro
	//[1038, 0x12b3], // Actris 1 Wireless
	//[1038, 0x12C2] // Arctis 9

	err := hid.Enumerate(0x1038, 0x1260, func(info *hid.DeviceInfo) error {
		steelHids = append(steelHids, info)

		return nil
	})
	if err != nil {
		log.Fatal("hid not found. Is it connected?")
	}

	for _, hidDevice := range steelHids {
		device, err := hid.OpenPath(hidDevice.Path)
		if err != nil {
			log.Println(fmt.Errorf("unable to connect to headset receiver err: %w", err))
		}

		_, err = getBattery(device)
		if err != nil {
			if errors.Is(err, errorNotConnected) {
				selectedHidDevice = device

				break
			}

			err = device.Close()
			if err != nil {
				log.Println(fmt.Errorf("err: %w", err))
			}

			continue
		} else {
			selectedHidDevice = device

			break
		}

	}
}

func reInit() {
	log.Println("reconnecting...")
	err := selectedHidDevice.Close()
	if err != nil {
		log.Println(fmt.Errorf("err: %w", err))
	}

	initHideDevice()
}
