package modules

import (
	"fmt"
	"time"

	"github.com/fatih/color"
)

func GetTime() string {
	currentTime := time.Now().Format("15:04:05")
	return color.BlueString(currentTime)
}

func Logger(status bool, content ...interface{}) {
	if status {
		fmt.Printf("[%s] %s", GetTime(), color.GreenString(fmt.Sprintln(content...)))
	} else if !status {
		fmt.Printf("[%s] %s", GetTime(), color.RedString(fmt.Sprintln(content...)))
	} else {
		fmt.Printf("[%s] %s", GetTime(), color.CyanString(fmt.Sprintln(content...)))
	}
}
