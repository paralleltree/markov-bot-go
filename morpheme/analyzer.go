package morpheme

type MorphemeAnalyzer interface {
	Analyze(sentence string) ([][]string, error)
}
