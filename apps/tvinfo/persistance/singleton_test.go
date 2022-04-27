package persistance

import (
	"os"
	"testing"
	"time"
)

func TestAppendAndContains(t *testing.T) {
	now := time.Now()
	err := Append(now.String())
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(persistanceFileName)

	contains, err := Contains(now.String())
	if err != nil {
		t.Fatal(err)
	}
	if !contains {
		t.Log("string should be contained")
		t.Fail()
	}
}
