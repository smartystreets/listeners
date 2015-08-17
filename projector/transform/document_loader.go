package transform

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/smartystreets/pipeline/projector/persist"
)

type DocumentLoader struct {
	client persist.HTTPClient
}

func NewDocumentLoader(client persist.HTTPClient) *DocumentLoader {
	return &DocumentLoader{client: client}
}

func (this *DocumentLoader) Load(path string, document interface{}) {
	request, err := http.NewRequest("GET", path, nil)
	if err != nil {
		log.Panic("Could not create request: ", err)
	}

	response, err := this.client.Do(request)
	if err != nil {
		log.Panic("HTTP Client Error: ", err)
	}

	if response.StatusCode == http.StatusNotFound {
		log.Printf("[INFO] Document not found at '%s'\n", path)
		return
	}

	if response.Body == nil {
		log.Panic("HTTP response body was nil")
	}

	defer response.Body.Close()

	reader := response.Body.(io.Reader)
	if response.Header.Get("Content-Encoding") == "gzip" {
		reader, _ = gzip.NewReader(reader)
	}

	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(document); err != nil {
		log.Panic("Document read error: ", err)
	}
}
