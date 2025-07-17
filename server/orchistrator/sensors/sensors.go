package sensors

import (
	"errors"
)

type Sensors struct {
	sensorType         *string
	registered_sensors []*sensor
	id_allowcations    []string
	ids_to_assign      []string
}

type sensor struct {
	hardwareId *string
	assignedId *string
}

type Sensors_interface interface {
	RegisterSensor(hardwareId string) (*sensor, error)
	DeregisterSensor(hardwareId string) (*sensor, error)
	UpdateAllocationList(newIdList []string) error
	SensorType() *string
}

var (
	ErrSensorAlreadyExists = "sensor already existis"
	ErrIdAlreadyAssigned   = "all id's are assigned"
	ErrFindingSensor       = "some error occured with finding the sensor in the array"
	ErrDuplicateIds        = "cannot have duplicate IDs in allocation list"
)

func New(sensorType *string, id_allowcations *[]string) (Sensors_interface, error) {
	if !validateIDAllocations(*id_allowcations) {
		return nil, errors.New(ErrDuplicateIds)
	}
	var sensors Sensors_interface = &Sensors{
		sensorType:      sensorType,
		id_allowcations: *id_allowcations,
		ids_to_assign:   *id_allowcations,
	}
	return sensors, nil
}

func (sensorArray *Sensors) SensorType() *string {
	return sensorArray.sensorType
}

func (sensorArray *Sensors) UpdateAllocationList(newIdList []string) error {
	if !validateIDAllocations(newIdList) {
		return errors.New(ErrDuplicateIds)
	}
	sensorArray.id_allowcations = newIdList
	sensorArray.ids_to_assign = newIdList
	var existingIdsToReRegister []string
	if len(sensorArray.registered_sensors) > 0 {
		for _, sensor := range sensorArray.registered_sensors {
			existingIdsToReRegister = append(existingIdsToReRegister, *sensor.hardwareId)
		}
		sensorArray.registered_sensors = make([]*sensor, 0)
		for _, id := range existingIdsToReRegister {
			_, err := sensorArray.RegisterSensor(id)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sensorArray *Sensors) RegisterSensor(hardwareId string) (*sensor, error) {

	if len(sensorArray.registered_sensors) > 0 {
		for _, registered_sensor := range sensorArray.registered_sensors {
			if *registered_sensor.hardwareId == hardwareId {
				return nil, errors.New(ErrSensorAlreadyExists)
			}
		}
	}

	if len(sensorArray.ids_to_assign) == 0 {
		return nil, errors.New(ErrIdAlreadyAssigned)
	}

	sensor := &sensor{
		hardwareId: &hardwareId,
		assignedId: &sensorArray.ids_to_assign[0],
	}
	sensorArray.registered_sensors = append(sensorArray.registered_sensors, sensor)
	sensorArray.ids_to_assign = sensorArray.ids_to_assign[1:]
	return sensor, nil
}

func (sensorArray *Sensors) DeregisterSensor(hardwareId string) (*sensor, error) {
	for index, registered_sensor := range sensorArray.registered_sensors {
		if *registered_sensor.hardwareId == hardwareId {
			sensorArray.registered_sensors = append(sensorArray.registered_sensors[:index], sensorArray.registered_sensors[index+1:]...)
			sensorArray.ids_to_assign = append(sensorArray.ids_to_assign, *registered_sensor.assignedId)
			return registered_sensor, nil
		}
	}
	return nil, errors.New(ErrFindingSensor)
}

func validateIDAllocations(slice []string) bool {
	seen := make(map[string]struct{}) // Use a struct{} for value as we only care about key presence

	for _, item := range slice {
		if _, exists := seen[item]; exists {
			return false // Duplicate found
		}
		seen[item] = struct{}{} // Mark item as seen
	}
	return true // No duplicates found
}
