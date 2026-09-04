package runner

import (
	"fmt"
	"os"
	"time"
)

func watchReloadFile(filepath string, reloadChan chan<- struct{}, duration time.Duration) {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for range ticker.C {
		_, err := os.Stat(filepath)
		if err == nil {
			fmt.Println("found reload file")
			reloadChan <- struct{}{}
			fmt.Println("sent channel message")
		}
	}
}
