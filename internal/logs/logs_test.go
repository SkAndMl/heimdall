package logs

import (
	"os"
	"reflect"
	"testing"
)

func TestGetLastNLinesIncludesUnterminatedFinalLine(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "log-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	if _, err := file.WriteString("first line\nlast line"); err != nil {
		t.Fatal(err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	got, err := getLastNLines(file, 0)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"first line\n", "last line"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("getLastNLines() = %#v, want %#v", got, want)
	}
}
