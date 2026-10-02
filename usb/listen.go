package usb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sstallion/go-hid"
)

type Listener struct {
	dev        *hid.Device
	packetSize int
	timeout    time.Duration
}

func NewListener(dev *hid.Device, packetSize int, timeout time.Duration) *Listener {
	if timeout <= 0 {
		timeout = 50 * time.Millisecond
	}
	return &Listener{
		dev:        dev,
		packetSize: packetSize,
		timeout:    timeout,
	}
}

func (l *Listener) Start(ctx context.Context) (<-chan []byte, <-chan error) {
	dataChan := make(chan []byte, 32)
	errChan := make(chan error, 1)

	go l.loop(ctx, dataChan, errChan)

	return dataChan, errChan
}

func (l *Listener) loop(ctx context.Context, dataChan chan<- []byte, errChan chan<- error) {
	defer close(dataChan)
	defer close(errChan)

	buf := make([]byte, l.packetSize)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := l.dev.ReadWithTimeout(buf, l.timeout)
			if err != nil {

				if errors.Is(err, hid.ErrTimeout) || err.Error() == "Interrupted system call" {
					continue
				}
				if ctx.Err() != nil {
					return
				}

				select {
				case errChan <- fmt.Errorf("hid read failed: %w", err):
				case <-ctx.Done():
				}
				return
			}

			if n == 0 {
				continue
			}

			packet := make([]byte, n)
			copy(packet, buf[:n])

			select {
			case dataChan <- packet:
			case <-ctx.Done():
				return
			}
		}
	}
}
