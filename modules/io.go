package modules

import (
	"bufio"
	"os"
)

func AppendLine(fileName, line string) {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	if _, err = file.WriteString(line + "\n"); err != nil {
		panic(err)
	}
}

func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func Contains(array []string, s interface{}) bool {
	for _, v := range array {
		if v == s {
			return true
		}
	}

	return false
}

func RemoveFromArray(array []string, s string) []string {
	var n []string
	for _, v := range array {
		if v != s {
			n = append(n, s)
		}
	}

	return n
}
