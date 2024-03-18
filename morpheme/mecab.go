package morpheme

import (
	"fmt"
	"io"
	"os/exec"
	"path"
	"strings"
)

type mecabAnalyzer struct {
	dicType string
}

func NewMecabAnalyzer(dicType string) *mecabAnalyzer {
	return &mecabAnalyzer{
		dicType: dicType,
	}
}

func (a *mecabAnalyzer) Analyze(text string) ([][]string, error) {
	preprocessed := PreprocessSentence(text)

	dicDir, err := resolveDicDir(a.dicType)
	if err != nil {
		return nil, fmt.Errorf("find dictionary dir: %w", err)
	}

	cmd := exec.Command("mecab", "-d", dicDir, "-Owakati")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("open stdin pipe: %w", err)
	}
	if _, err := io.WriteString(stdin, preprocessed); err != nil {
		return nil, fmt.Errorf("write to stdin: %w", err)
	}
	if err := stdin.Close(); err != nil {
		return nil, fmt.Errorf("close stdin: %w", err)
	}
	bytes, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("run mecab: %w", err)
	}

	sentences := strings.Split(string(bytes), "\n")
	res := make([][]string, 0, len(sentences))
	for _, sentence := range sentences {
		if len(sentence) == 0 {
			continue
		}
		words := strings.Split(sentence, " ")
		filtered := make([]string, 0, len(words))
		for _, v := range words {
			if v == "" {
				continue
			}
			filtered = append(filtered, v)
		}
		res = append(res, filtered)
	}

	return res, nil
}

func resolveDicDir(dicType string) (string, error) {
	dicDir, err := exec.Command("mecab-config", "--dicdir").Output()
	if err != nil {
		return "", fmt.Errorf("read mecab stdout: %w", err)
	}
	return path.Join(strings.TrimSuffix(string(dicDir), "\n"), dicType), nil
}
