package main

import "fmt"
import "time"

type config struct {
	startTime time.Time
	endTime   time.Time
}

func main() {
	conf := config{}
	conf.startTime = time.Now()
	conf.endTime = (conf.startTime).Add(time.Minute * time.Duration(60))

	fmt.Println("Start time:", conf.startTime)
	fmt.Println("End time:", conf.endTime)
}
