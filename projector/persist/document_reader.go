package persist

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type DocumentReader struct {
	client HTTPClient
}

func NewDocumentReader(client HTTPClient) *DocumentReader {
	return &DocumentReader{client: client}
}

func (this *DocumentReader) Read(path string, document interface{}) error {
	request, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return fmt.Errorf("Could not create request: '%s'", err.Error())
	}

	response, err := this.client.Do(request)
	if err != nil {
		return fmt.Errorf("HTTP Client Error: '%s'", err.Error())
	}

	if response.StatusCode == http.StatusNotFound {
		log.Printf("[INFO] Document not found at '%s'\n", path)
		return nil
	}

	if response.Body == nil {
		return fmt.Errorf("HTTP response body was nil: '%s'", err.Error())
	}

	defer response.Body.Close()

	reader := response.Body.(io.Reader)
	if response.Header.Get("Content-Encoding") == "gzip" {
		reader, _ = gzip.NewReader(reader)
	}

	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(document); err != nil {
		return fmt.Errorf("Document read error: '%s'", err.Error())
	}

	return nil
}

func (this *DocumentReader) ReadPanic(path string, document interface{}) {
	if err := this.Read(path, document); err != nil {
		log.Panic(err)
	}
}