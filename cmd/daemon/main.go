package main

import (
	"context"
	"fmt"
	"log"
	"mixeryd/models"
	"mixeryd/usb"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/sstallion/go-hid"
)

var apps = map[byte]string{
	1: "ghostty",
	2: "firefox",
	3: "thunar",
}

const (
	MIXERY_VID = 0xCAFE
	MIXERY_PID = 0x4004
)

func main() {

	if err := hid.Init(); err != nil {
		log.Fatal(err)
	}

	d, err := hid.OpenFirst(MIXERY_VID, MIXERY_PID)
	if err != nil {
		log.Fatal(err)
	}

	s, err := d.GetMfrStr()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Manufacturer String: %s\n", s)

	// Read the Product String.
	s, err = d.GetProductStr()
	if err != nil {
		log.Fatal(err)
	}
	l := usb.NewListener(d, models.MaxPacketSize, 50*time.Millisecond)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	dataChan, errChan := l.Start(ctx)
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n[Ctrl+C] Signal received!")
			return

		case err, ok := <-errChan:
			if !ok {
				errChan = nil
				continue
			}
			log.Fatalln("listener error:", err)

		case data, ok := <-dataChan:
			if !ok {
				dataChan = nil
				continue
			}

			slot := models.DeviceSlot{}

			soe, err := slot.Parse(data)
			if err != nil {
				log.Fatalln("parsing error:", err)
				continue
			}
			switch soe {
			case models.REPORT_ID_INPUT_BUTTON:
				btn, err := slot.GetButton()
				if err != nil {
					log.Fatalln("parsing error:", err)
					continue
				}
				if slot.Action.Type == models.ACTION_LAUNCH && btn.Pressed == 1 {
					app := apps[slot.Action.Id]
					exec.Command(app).Start()

				}
				log.Printf("%+v\n", slot)
			case models.REPORT_ID_INPUT_ENCODER:
				enc, err := slot.GetEncoder()
				if err != nil {
					log.Fatalln("parsing error:", err)
					continue
				}
				fmt.Printf("enc: %v\n", enc)
				log.Printf("%+v\n", slot)

			}

		}
	}
}
