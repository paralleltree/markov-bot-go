package morpheme

import (
	"regexp"
	"strings"
)

func PreprocessSentence(sentence string) string {
	if regexp.MustCompile("https?://").MatchString(sentence) {
		return ""
	}
	replacer := strings.NewReplacer(
		"!", "！",
		"?", "？",
		"，", "、",
		"．", "。",
		"。", "。\n",
		"&lt;", "<",
		"&gt;", ">",
		"&amp;", "&",
		"&nbsp;", "",
	)
	return replacer.Replace(sentence)
}
