package flags

func levenshtein(s string, t string) int {
	// Work on runes rather than bytes, otherwise multi-byte characters are
	// each counted as several independent edits.
	sr := []rune(s)
	tr := []rune(t)

	if len(sr) == 0 {
		return len(tr)
	}

	if len(tr) == 0 {
		return len(sr)
	}

	// Only the previous and the current row of the distance matrix are ever
	// needed at the same time.
	prev := make([]int, len(tr)+1)
	cur := make([]int, len(tr)+1)

	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(sr); i++ {
		cur[0] = i

		for j := 1; j <= len(tr); j++ {
			if sr[i-1] == tr[j-1] {
				cur[j] = prev[j-1]
				continue
			}

			// Substitution, insertion or deletion, whichever is cheapest.
			cur[j] = min(prev[j-1], cur[j-1], prev[j]) + 1
		}

		prev, cur = cur, prev
	}

	return prev[len(tr)]
}

func closestChoice(cmd string, choices []string) (string, int) {
	if len(choices) == 0 {
		return "", 0
	}

	mincmd := -1
	mindist := -1

	for i, c := range choices {
		l := levenshtein(cmd, c)

		if mincmd < 0 || l < mindist {
			mindist = l
			mincmd = i
		}
	}

	return choices[mincmd], mindist
}
