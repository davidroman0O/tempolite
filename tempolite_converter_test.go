package tempolite

import (
	"fmt"
	"reflect"
	"testing"
)

type CustomId string

func (s CustomId) String() string {
	return string(s)
}
func TestCustomType(t *testing.T) {
	rawInput := "some_spotify_id"
	desiredType := reflect.TypeOf(CustomId(""))
	desiredKind := desiredType.Kind()

	convertedValue, err := convertIO(rawInput, desiredType, desiredKind)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fail()
	} else {
		scustomID := convertedValue.(CustomId)
		fmt.Println("Converted scustomID:", scustomID)
	}
}
