package til

import (
	"os"
	"strconv"
	"testing"
)

var Debug bool

func setup() bool {
	debug, err := strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		debug = false
	}
	return debug
}

func TestMain(m *testing.M) {
	Debug = setup()

	code := m.Run()
	os.Exit(code)
}
