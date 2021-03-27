package bus

import (
	"fmt"
	"io"
	"runtime"
	"time"

	"github.com/jacobsa/go-serial/serial"
	e "github.com/ledboard/pkg/errors"
)

const asciiByteOffset = 48

type SerialBus struct {
	options serial.OpenOptions
	bus     io.ReadWriteCloser
	readCh  chan uint8
	writeCh chan LedLight
	log     *e.Logger
}

type LedLight struct {
	Num   uint8
	State bool
}

func (s *SerialBus) runRead() uint8 {
	for {
		buf := make([]byte, 1)
		s.bus.Read(buf)
		s.readCh <- buf[0]
	}
}

func (s *SerialBus) runWrite() {
	var newLedState LedLight
	var state []byte
	for {
		newLedState = <-s.writeCh
		switch {
		case newLedState.State:
			state = []byte("H")
		default:
			state = []byte("L")
		}
		ledNum := newLedState.Num + asciiByteOffset
		n, err := s.bus.Write(append([]byte{ledNum}, state...))
		s.log.Debugf("sentBytes:%v, len:%v", append([]byte{ledNum}, state...), n)
		if err != nil {
			s.log.Errorf("err:%v", err.Error())
		}
	}
}

func (s *SerialBus) Connect() {
	port, err := serial.Open(s.options)
	s.log.CheckPanic(err)
	buf := make([]byte, 1)
	_, err = port.Read(buf)
	s.log.CheckPanic(err)
	if string(buf[0]) != "s" {
		panic(fmt.Sprintf("error reading from ledboard. got %v", buf[0]))
	}
	_, err = port.Write([]byte{0})
	s.log.CheckPanic(err)
	time.Sleep(time.Second)
	s.log.Info("connection successful")
	s.bus = port
	s.readCh = make(chan uint8, 256)
	s.writeCh = make(chan LedLight, 256)
	go s.runRead()
	go s.runWrite()
}
func (s *SerialBus) Disconnect() {
	s.bus.Close()
}

func (s *SerialBus) GetReadChannel() chan uint8 {
	return s.readCh
}

func (s *SerialBus) GetWriteChannel() chan LedLight {
	return s.writeCh
}

func NewSerialBus(port string, baud uint, logger *e.Logger) SerialBus {
	bus := SerialBus{
		options: serial.OpenOptions{
			PortName: port,
			BaudRate: baud,
			DataBits: 8,
		},
		log: logger,
	}
	if runtime.GOOS != "windows" {
		bus.options.StopBits = 1
		bus.options.MinimumReadSize = 1
	}
	return bus
}
