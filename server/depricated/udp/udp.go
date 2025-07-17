package udp

import (
	"errors"
	"fmt"
	"net"
	"slices"
)

type UDP struct {
	localAddr  *net.UDPAddr
	remoteAddr *net.Addr
	Conn       *net.UDPConn
	Buf        []byte
}

type SensorDatagram struct {
	Length              int
	Version             int
	HardwareId          int
	RequestedSoftwareId int
}

const (
	ErrInvalidPairRequest = 45
)

const (
	version = 1
)

func NewUDP() *UDP {
	udp := &UDP{
		localAddr: &net.UDPAddr{
			IP:   net.IPv4(192, 168, 10, 188),
			Port: 3333,
			Zone: "",
		},
	}
	return udp
}

func (udp *UDP) SetupConnection() error {
	conn, err := net.ListenUDP("udp", udp.localAddr)
	if err != nil {
		//TODO add err handling
		return err
	}
	udp.Conn = conn
	return nil
}

func (udp *UDP) GetBufferFromSender() error {
	buf := make([]byte, 4)
	_, remoteAddr, err := udp.Conn.ReadFrom(buf)
	if err != nil {
		println(err)
		//TODO add error handling
		return err
	}
	udp.Buf = buf
	fmt.Println(remoteAddr)
	udp.remoteAddr = &remoteAddr
	return nil
}

func (udp *UDP) DecodeMessage() (*SensorDatagram, error) {
	datagram := &SensorDatagram{
		Length:              int(udp.Buf[0]),
		Version:             int(udp.Buf[1]),
		HardwareId:          int(udp.Buf[2]),
		RequestedSoftwareId: int(udp.Buf[3]),
	}
	if datagram.Version != 1 {
		return nil, errors.New("invalid payload")
	}
	return datagram, nil
}

func (udp *UDP) SendResponse(errMessage byte) error {
	var err error
	if errMessage == 0 {
		err = udp.sendSuccessResponse()
	} else {
		err = udp.sendFailResponse(errMessage)
	}
	if err != nil {
		fmt.Printf("Error sending response: %v\n", err)
	}
	return err
}

func (udp *UDP) sendSuccessResponse() error {
	response := []byte{1, version}
	err := udp.respond(response)
	return err
}

func (udp *UDP) sendFailResponse(failCode byte) error {
	response := []byte{failCode, 0, version}
	err := udp.respond(response)
	return err
}

func (udp *UDP) respond(response []byte) error {
	response = append(response, byte(len(response)))
	slices.Reverse(response)
	_, err := udp.Conn.WriteTo(response, *udp.remoteAddr)
	return err
}

//Pairing handshake:
// eye send request to pair
//server assigns given software id to eye and send successful back
// eye send sensor data

//if server send unsuccesful pair, eye will retry 3 times, then wait 5s then retry again, then wait 1min then retry again, then wait 5min and retry.

//initial buffer data format from eye: byte 1 = length of message, byte 2 = version, byte 3 = hardware id, byte 4 = requested software id
//response to connection request from server format: byte 1 = length of message, byte 2 = version, byte 3 = successful pair bool, byte 4 = fail code(optional)
//request from eye after pair successful: byte 1 = length of message, byte 2 = version, byte 3 = hardware id
