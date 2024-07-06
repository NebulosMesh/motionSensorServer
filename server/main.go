package main

import (
	"errors"
	"fmt"

	EYES "github.com/superbrobenji/motionServer/sensors"
	UDP "github.com/superbrobenji/motionServer/udp"

	"log"
)

func main() {
	println("Starting up brain")
	udp := UDP.NewUDP()
	println("Setting up new connection to eyes")
	err := udp.SetupConnection()
	if err != nil {
		log.Fatal(err)
	}
	println("Creating eyes")
	eyes := EYES.NewEyes()
	defer udp.Conn.Close()

	println("Listenting for eyes")
	for {
		sensorLoopHandler(udp, eyes)
	}
}

func sensorLoopHandler(udp *UDP.UDP, eyes *EYES.Eyes) error {
	err := udp.GetBufferFromSender()
	if err != nil {
		//TODO error handling
		return errors.New("nothing to read")
	}
	message, err := udp.DecodeMessage()
	if err != nil {
		fmt.Printf("failed to decode message: %v", err)
	}

	currentEye := eyes.GetEyeFromHardwareId(message.HardwareId)

	var routineErr error

	go eyeCreationHandler(currentEye, message, eyes, udp, &routineErr)

	go existingEyeHandler(currentEye, message, udp, &routineErr)
	if routineErr != nil {
		return routineErr
	}
	return nil
}

func eyeCreationHandler(currentEye *EYES.Eye, message *UDP.SensorDatagram, eyes *EYES.Eyes, udp *UDP.UDP, routineErr *error) {
	if *currentEye == (EYES.Eye{}) {
		err := eyes.CreateEye(message)
		if err != nil {
			udp.SendResponse(UDP.ErrInvalidPairRequest)
			fmt.Printf("err: %v\n", err)
			*routineErr = errors.New("failed to create eye")
		} else {
			udp.SendResponse(0)
			println("new eye created")
		}
	}
}

func existingEyeHandler(currentEye *EYES.Eye, message *UDP.SensorDatagram, udp *UDP.UDP, routineErr *error) {
	if *currentEye != (EYES.Eye{}) {
		isReset := resetPair(message, currentEye, udp)
		if isReset == true {
			udp.SendResponse(0)
			println("reset pair")
		} else {
			err := processEyeData(currentEye)
			//TODO add normal behaviour here
			*routineErr = err
		}
	}
}

func resetPair(message *UDP.SensorDatagram, currentEye *EYES.Eye, udp *UDP.UDP) bool {
	if message.RequestedSoftwareId != 0 {
		//TODO decide if duplicate software keys are allowed. if not add check here
		if currentEye.MappedId != message.RequestedSoftwareId {
			currentEye.MappedId = message.RequestedSoftwareId
			if currentEye.Paired == false {
				currentEye.Paired = true
			}
			println("pair reset")
			udp.SendResponse(0)
		}
		return true
	}
	return false
}

func processEyeData(currentEye *EYES.Eye) error {
	fmt.Printf("normal behaviour after pair for eye: %v\n", currentEye)
	return nil
}
