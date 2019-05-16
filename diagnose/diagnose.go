package diagnose

import (
	"fmt"
	"os"
)

func Diagnose() {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = err.Error()
	}
	fmt.Printf("Diagnosing %s\n", hostname)
}
