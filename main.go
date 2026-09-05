package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

type Word struct {
	word   string
	length int
	repets int
}

func main() {

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Ошибка: Не указано имя текстового файла")
		os.Exit(1)
	}

	filePath := os.Args

	file, err := os.Open(filePath[1])

	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка открытия файла: %v", err)
		os.Exit(1)
	}

	wordList := make([]Word, 0)
	wordMap := make(map[string]int)

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	text := ""
	lineCount := 0

	for scanner.Scan() {
		lineCount++
		text += scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Ошибка при сканировании: %v\n", err)
		os.Exit(1)
	}
}
