package rag

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
)

// ExtractText dispatches to the correct extractor based on file extension.
func ExtractText(path, ext string) (string, error) {
	log.Printf("[EXTRACT] file=%s type=%s", filepath.Base(path), ext)

	var (
		text string
		err  error
	)

	switch strings.ToLower(ext) {
	case ".pdf":
		text, err = extractPDF(path)
	case ".txt", ".md", ".csv", ".json", ".xml", ".html", ".htm":
		text, err = extractPlainText(path)
	case ".docx", ".doc":
		text, err = extractDOCX(path)
	default:
		return "", fmt.Errorf("no extractor for extension %q", ext)
	}

	if err != nil {
		log.Printf("[EXTRACT] ERROR file=%s: %v", filepath.Base(path), err)
		return "", err
	}

	log.Printf("[EXTRACT] file=%s extracted_chars=%d", filepath.Base(path), len(text))
	return text, nil
}

// extractPDF reads a PDF using ledongthuc/pdf.
func extractPDF(path string) (string, error) {
	_, r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("pdf.Open: %w", err)
	}

	b, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("GetPlainText: %w", err)
	}

	var sb strings.Builder
	buf := make([]byte, 4096)
	for {
		n, readErr := b.Read(buf)
		sb.Write(buf[:n])
		if readErr != nil {
			break
		}
	}
	return sb.String(), nil
}

// extractPlainText reads the file as-is (TXT, MD, CSV, JSON, XML, HTML).
func extractPlainText(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("os.ReadFile: %w", err)
	}
	return string(data), nil
}

// extractDOCX unzips a .docx and pulls text from word/document.xml.
func extractDOCX(path string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", fmt.Errorf("zip.OpenReader: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name != "word/document.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", fmt.Errorf("opening word/document.xml: %w", err)
		}
		defer rc.Close()

		raw, err := io.ReadAll(rc)
		if err != nil {
			return "", fmt.Errorf("reading word/document.xml: %w", err)
		}
		return stripXMLTags(raw), nil
	}
	return "", fmt.Errorf("word/document.xml not found in archive")
}

// stripXMLTags decodes XML and concatenates all character data.
func stripXMLTags(data []byte) string {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var sb strings.Builder
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		if cd, ok := tok.(xml.CharData); ok {
			text := strings.TrimSpace(string(cd))
			if text != "" {
				sb.WriteString(text)
				sb.WriteRune(' ')
			}
		}
	}
	return sb.String()
}
