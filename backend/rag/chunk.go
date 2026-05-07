package rag

func Chunk(text string, size int) []string {
	var out []string

	for len(text) > 0 {
		if len(text) < size {
			out = append(out, text)
			break
		}
		out = append(out, text[:size])
		text = text[size:]
	}

	return out
}
