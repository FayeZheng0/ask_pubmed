package pubmed

import (
	"sort"

	"github.com/FayeZheng0/ask_pubmed/core"
)

func OptimizeKeywords(documentInfo []core.Paper) []string {
	keywordCounter := make(map[string]float32)
	for _, doc := range documentInfo {
		for _, kw := range doc.Keywords {
			keywordCounter[kw] += doc.SearchScore
		}
	}

	candidateKeywords := getTopK(keywordCounter)

	optimizedKeywords := removeDuplicates(candidateKeywords)

	return optimizedKeywords
}

func getTopK(m map[string]float32) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return m[keys[i]] > m[keys[j]]
	})
	if len(keys) > 3 {
		keys = keys[:3]
	}
	return keys
}

func removeDuplicates(strs []string) []string {
	var result []string
	seen := make(map[string]bool)
	for _, str := range strs {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}
	return result
}
