package sensors

import (
	"testing"
)

func TestRegisterSensor(t *testing.T) {
	id_allowances := []string{"1", "2", "3", "4"}
	topic := "testcases"
	sensors, _ := New(&topic, &id_allowances)

	var tests = []struct {
		name  string
		input string
		want  string
	}{
		{"should create a sensor", "2", "1"},
		{"should throw error that sensor already exists", "2", ErrSensorAlreadyExists},
		{"should create a sensor", "1", "2"},
		{"should create a sensor", "8", "3"},
		{"should create a sensor", "7", "4"},
		{"should throw err that all allowed ids are used", "5", ErrIdAlreadyAssigned},
	}
	// The execution loop
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans, err := sensors.RegisterSensor(tt.input)
			if err != nil {
				if err.Error() != tt.want {
					t.Errorf("got %s, want %s", err, tt.want)
				}
			} else {
				if *ans.assignedId != tt.want {
					t.Errorf("got %v\n, want %s", ans.assignedId, tt.want)
				}
				if *ans.hardwareId != tt.input {
					t.Errorf("got %v\n, want %s", ans.hardwareId, tt.input)
				}
			}

		})
	}
}

func TestDeregisterSensor(t *testing.T) {
	id_allowances := []string{"1", "2", "3", "4"}
	ids := []string{"2", "3", "4", "5"}
	topic := "testcases"
	sensors, _ := New(&topic, &id_allowances)
	for _, id := range ids {
		_, err := sensors.RegisterSensor(id)
		if err != nil {
			t.Errorf("got %s, want %s", err, "no error")
		}
	}

	var tests = []struct {
		name  string
		input string
		want  string
	}{
		{"should remove a sensor", "2", "1"},
		{"should remove a sensor", "8", ErrFindingSensor},
		{"should remove a sensor", "3", "2"},
		{"should remove a sensor", "4", "3"},
		{"should remove a sensor", "5", "4"},
	}
	// The execution loop
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans, err := sensors.DeregisterSensor(tt.input)
			if err != nil {
				if err.Error() != tt.want {
					t.Errorf("got %s, want %s", err, tt.want)
				}
			} else {
				if *ans.assignedId != tt.want {
					t.Errorf("got %v\n, want %s", ans.assignedId, tt.want)
				}
				if *ans.hardwareId != tt.input {
					t.Errorf("got %v\n, want %s", ans.hardwareId, tt.input)
				}

			}

		})
	}
}

func TestUpdateIdAllocation(t *testing.T) {
	id_allowances := []string{"1", "2", "3", "4"}
	new_id_allowances := []string{"8", "9", "10", "11"}
	ids := []string{"2", "3", "4", "5"}
	topic := "testcases"
	sensors, _ := New(&topic, &id_allowances)
	for _, id := range ids {
		_, err := sensors.RegisterSensor(id)
		if err != nil {
			t.Errorf("got %s, want %s", err, "no error")
		}
	}
	err := sensors.UpdateAllocationList(new_id_allowances)
	if err != nil {
		t.Errorf("got %s, want %s", err, "no error")
	}
}

func TestNoDuplicates(t *testing.T) {
	id_wrong_allowances := []string{"1", "1", "3", "4"}
	id_allowances := []string{"1", "2", "3", "4"}
	new_id_wrong_allowances := []string{"8", "8", "10", "11"}
	ids := []string{"2", "3", "4", "5"}
	topic := "testcases"

	sensors, err := New(&topic, &id_wrong_allowances)
	if err == nil {
		t.Errorf("got %s, want %s", err, ErrDuplicateIds)
	}

	sensors, err = New(&topic, &id_allowances)
	for _, id := range ids {
		_, err := sensors.RegisterSensor(id)
		if err != nil {
			t.Errorf("got %s, want %s", err, "no error")
		}
	}

	err = sensors.UpdateAllocationList(new_id_wrong_allowances)
	if err == nil {
		t.Errorf("got %s, want %s", err, ErrDuplicateIds)
	}
}
