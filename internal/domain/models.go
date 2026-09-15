package domain

type Word struct {
	Text     string
	Length   int
	Quantity int
}

type AnalysisResult struct {
	FilePath      string  `json:"file-path"`
	Lines         int     `json:"lines"`
	Symbols       int     `json:"symbols"`
	Sentences     int     `json:"sentences"`
	WordCount     int     `json:"words"`
	AvgWordLength float64 `json:"avg-word-len"`
	LongestWord   Word    `json:"longest-word"`
	TopFrequents  []*Word `json:"top-frequents"`
}
