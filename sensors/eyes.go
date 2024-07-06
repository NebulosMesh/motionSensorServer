package eyes

import (
	"errors"
	UDP "github.com/superbrobenji/motionServer/udp"
	"slices"
)

const maxEyes = 10

type Eyes struct {
	slice []Eye
}

type Eye struct {
	HardwareId int
	MappedId   int
	Paired     bool
}

func NewEyes() *Eyes {
	return &Eyes{
		slice: make([]Eye, maxEyes),
	}
}

func (eyes *Eyes) GetEyeFromHardwareId(id int) *Eye {
	index := slices.IndexFunc(eyes.slice, func(e Eye) bool { return e.HardwareId == id })
	if index < 0 {
		return &Eye{}
	}
	return &eyes.slice[index]
}

func (eyes *Eyes) CreateEye(message *UDP.SensorDatagram) error {
	if message.RequestedSoftwareId == 0 {
		return errors.New("invalid request")
	}
	eyes.createNewStructEye(message)
	return nil
}

func (eyes *Eyes) createNewStructEye(message *UDP.SensorDatagram) {
	eye := &Eye{
		HardwareId: message.HardwareId,
		MappedId:   message.RequestedSoftwareId,
		Paired:     true,
	}
	index := eyes.findNextEmptyEye()
	eyes.slice[index] = *eye
}

func (eyes *Eyes) findNextEmptyEye() int {
	return slices.IndexFunc(eyes.slice, func(e Eye) bool { return e == (Eye{}) })
}

//if sensor exists, check if softwareId is provided, if it is, change software id on eye. respond of successful pair
//if sensor does not exist, set hardware and software id. respond with successful pair.
//if sensor does not exist and no software key provided. respond with unsuccessful pair.
//if sensor exist and no software key provided. do not respond
