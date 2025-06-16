package bars

import (
	"cmp"
	"slices"
	"strconv"
	"strings"
)

func Bars(m map[string]int) string {
	if len(m) == 0 {
		return ""
	}

	type pair struct {
		Key   string
		Value int
	}

	pairs := make([]pair, 0, len(m))
	var keyLength int
	for k, v := range m {
		pairs = append(pairs, pair{Key: k, Value: v})
		keyLength = max(keyLength, len(k))
	}

	slices.SortFunc(pairs, func(a, b pair) int {
		return -cmp.Compare(a.Value, b.Value)
	})

	var sb strings.Builder
	for _, p := range pairs {
		sb.WriteString(p.Key)
		for range keyLength - len(p.Key) {
			sb.WriteString(" ")
		}
		sb.WriteString("\t")
		sb.WriteString(strconv.Itoa(p.Value))
		sb.WriteString("\t")
		for range p.Value {
			sb.WriteString(".")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
