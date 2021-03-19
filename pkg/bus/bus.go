package bus

import (
	"fmt"
	"io"
	"time"

	"github.com/jacobsa/go-serial/serial"
	e "github.com/ledboard/pkg/errors"
)

const asciiByteOffset = 48

type SerialBus interface {
	Connect()
	GetReadChannel() chan uint8
	GetWriteChannel() chan LedLight
}

type serialBus struct {
	options serial.OpenOptions
	bus     io.ReadWriteCloser
	readCh  chan uint8
	writeCh chan LedLight
}

type LedLight struct {
	Num   uint8
	State bool
}

func (s *serialBus) runRead() uint8 {
	for {
		buf := make([]byte, 1)
		s.bus.Read(buf)
		s.readCh <- buf[0]
	}
}

func (s *serialBus) runWrite() {
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
		fmt.Printf("sent:%v, len:%v\n", n, append([]byte{ledNum}, state...))
		if err != nil {
			fmt.Printf("err:%v\n", err.Error())
		}
	}
}

func (s *serialBus) Connect() {
	port, err := serial.Open(s.options)
	e.CheckPanic(err)
	buf := make([]byte, 1)
	_, err = port.Read(buf)
	e.CheckPanic(err)
	if string(buf[0]) != "s" {
		panic("error reading from ledboard")
	}
	_, err = port.Write([]byte{0})
	e.CheckPanic(err)
	time.Sleep(time.Second)
	fmt.Println("connection successful")

	s.bus = port
	s.readCh = make(chan uint8, 256)
	s.writeCh = make(chan LedLight, 256)
	go s.runRead()
	go s.runWrite()
}

func (s *serialBus) GetReadChannel() chan uint8 {
	return s.readCh
}

func (s *serialBus) GetWriteChannel() chan LedLight {
	return s.writeCh
}

func NewSerialBus(port string, baud uint) SerialBus {
	bus := &serialBus{
		options: serial.OpenOptions{
			PortName: port,
			BaudRate: baud,
			DataBits: 8,
		},
	}
	return bus
}
