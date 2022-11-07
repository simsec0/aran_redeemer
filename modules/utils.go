package modules

import (
	"os"
	"os/exec"
)

func Clear() {
	cmd := exec.Command("cmd", "/C", "cls || clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func SetTitle(title string) {
	cmd := exec.Command("cmd", "/C", "title", title)
	cmd.Stdout = os.Stdout
	cmd.Run()
}
