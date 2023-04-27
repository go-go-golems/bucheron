package pkg

import (
	_ "embed"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"strings"
)

//go:embed words.txt
var wordsFile string
var words []string
var wordIndex map[string]int

func init() {
	words = strings.Split(wordsFile, "\n")
	if len(words) != 65536 {
		panic(fmt.Sprintf("words.txt must contain exactly 65536 words, contained: %d", len(words)))
	}
	wordIndex = make(map[string]int, 65536)
	for i, word := range words {
		wordIndex[word] = i
	}
}

func UUIDToHorseStaple(id uuid.UUID) string {
	var ret []string
	for i := 0; i < 16; i += 2 {
		idx := uint16(id[i])<<8 | uint16(id[i+1])
		word := words[idx]
		ret = append(ret, word)
	}

	return strings.Join(ret, "-")
}

func HorseStapleToUUID(id string) (uuid.UUID, error) {
	var ret uuid.UUID
	for i, word := range strings.Split(id, "-") {
		idx, ok := wordIndex[word]
		if !ok {
			return ret, errors.Errorf("invalid horse staple: %s, invalid word: %s", id, word)
		}
		ret[i*2] = byte(idx >> 8)
		ret[i*2+1] = byte(idx)
	}

	return ret, nil
}
