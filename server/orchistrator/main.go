package main

import (
	"fmt"

	EventStore "github.com/superbrobenji/motionServer/eventStore"
	Sensors "github.com/superbrobenji/motionServer/sensors"
)

var (
	sensorType          = "motion"
	broker              = "localhost:9092"
	groupId             = "1"
	topic               = "motion-trigger"
	sensorIDAllocations = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
)

func main() {
	//setup store
	eventStore := EventStore.New(broker, groupId)
	err := eventStore.Connect()
	if err != nil {
		panic(err)
	}

	//todo creage a web server for configs. to do all this
	//todo create functions to set up different sensors
	motionSensors, err := Sensors.New(&sensorType, &sensorIDAllocations)
	if err != nil {
		panic(err)
	}
	//remove
	fmt.Print(motionSensors)

	//write to store
	// err = eventStore.WriteMessage("hello, world", topic)
	// if err != nil {
	// 	panic(err)
	// }
	//subscripe to topic
	// err = eventStore.SubscribeToEvents(topic)
	// if err != nil {
	// 	panic(err)
	// }

}
