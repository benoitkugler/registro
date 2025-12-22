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

func TestMarshalDate(t *testing.T) {
	d := randDate()
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var d2 Date
	if err := json.Unmarshal(b, &d2); err != nil {
		t.Fatal(err)
	}
	if !d2.Equals(d) {
		t.Fatalf("unexpected dates : %v %v", d, d2)
	}
}
