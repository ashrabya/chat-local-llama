package rag

import (
	"github.com/ledongthuc/pdf"
)

func ExtractPDF(path string) (string, error) {
	_, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}

	var text string

	b, _ := r.GetPlainText()
	buf := make([]byte, 1024)

	for {
		n, err := b.Read(buf)
		text += string(buf[:n])
		if err != nil {
			break
		}
	}

	return text, nil
}
