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

	pattern := `[,;:.!?*#]+`
	re := regexp.MustCompile(pattern)

	regText := re.Split(text, -1)

	patternSugg := `[.!?]+`
	reSugg := regexp.MustCompile(patternSugg)
	textSugg := reSugg.Split(text, -1)

	var words []string

	for _, t := range regText {
		if t != "" {

			words = append(words, strings.Fields(t)...)
		}
	}

	for _, w := range words {

		if idx, ok := wordMap[w]; ok {
			wordList[idx].repets++
		} else {
			newIdx := len(wordList)
			wordMap[w] = newIdx
			wordList = append(wordList, Word{word: w, length: utf8.RuneCountInString(w), repets: 1})
		}
	}

	fmt.Printf("Статистика текста (%s):\n", file.Name())
	fmt.Printf("Символов: %d\n", countSymbols(text))
	fmt.Printf("Строк: %d\n", lineCount)
	fmt.Printf("Предложений: %d\n", countSuggest(textSugg))
	fmt.Printf("Слов: %d\n", sumRepets(wordList))
	fmt.Printf("Средняя длина слова: %v символов\n", avgWordLen(wordList))

	lw := longestWord(wordList)
	fmt.Printf("Самое длинное слово: \"%s\" (%d символов)\n", lw.word, lw.length)

	fmt.Println("\nТоп-5 самых частых слов: ")
	printWordList(frequentsWord(wordList, 5))

	fmt.Printf("\nDone!\n")
}

func countSymbols(text string) int {
	return utf8.RuneCountInString(text)
}

func printWordList(wordList []Word) {

	strQty := ""

	for i, w := range wordList {
		if w.repets%10 == 2 || w.repets%10 == 3 || w.repets%10 == 4 {
			strQty = "раза"
		} else {
			strQty = "раз"
		}

		fmt.Printf("%d. \"%s\" — %d %s\n", i+1, w.word, w.repets, strQty)
	}
}

func avgWordLen(wordList []Word) int {

	sum := 0
	countWords := 0

	for _, w := range wordList {

		sum += w.length * w.repets
		countWords += w.repets
	}
	return int(math.Ceil(float64(sum) / float64(countWords)))
}

func longestWord(wordList []Word) Word {
	sort.Slice(wordList, func(i, j int) bool {
		return wordList[i].length > wordList[j].length
	})

	return wordList[0]
}

func frequentsWord(wordList []Word, qty int) []Word {
	sort.Slice(wordList, func(i, j int) bool {
		return wordList[i].repets > wordList[j].repets
	})

	var topWords []Word

	for i := 0; i < qty; i++ {
		topWords = append(topWords, wordList[i])
	}

	return topWords
}

func sumRepets(wordList []Word) int {
	var total int
	for _, w := range wordList {
		total += w.repets
	}
	return total
}

func countSuggest(text []string) int {
	count := 0

	for _, t := range text {
		if t != "" {
			count++
		}
	}

	return count
}
