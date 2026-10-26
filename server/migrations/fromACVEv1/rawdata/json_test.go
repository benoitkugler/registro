package rawdata

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	var ti time.Time
	fmt.Println(ti.Nanosecond())
	b, _ := json.Marshal(ti)
	fmt.Println(string(b))

	var ti2 time.Time
	if err := json.Unmarshal(b, &ti2); err != nil {
		t.Fatal(err)
	}
	fmt.Println(ti2)
}
