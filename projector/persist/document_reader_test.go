package persist

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/smartystreets/assertions/should"
	"github.com/smartystreets/gunit"
)

type DocumentReaderFixture struct {
	*gunit.Fixture

	path     string
	reader   *DocumentReader
	client   *FakeHTTPGetClient // HTTPClient
	stdout   *bytes.Buffer
	document *Document
}

func (this *DocumentReaderFixture) Setup() {
	this.path = "/document/path"
	this.client = &FakeHTTPGetClient{}
	this.reader = NewDocumentReader(this.client)
	this.document = &Document{}
	this.stdout = new(bytes.Buffer)
	log.SetOutput(this.stdout)
}

func (this *DocumentReaderFixture) TestRequestInvalid_ClientIgnored() {
	this.path = "%%%%%%%%"
	this.assertPanic(`Could not create request: parse %%%%%%%%: invalid URL escape "%%%"`)
	this.So(this.client.called, should.BeFalse)
}

func (this *DocumentReaderFixture) TestClientErrorPreventsDocumentReading() {
	this.client.err = errors.New("BOINK!")
	this.assertPanic("HTTP Client Error: BOINK!")
}

func (this *DocumentReaderFixture) TestDocumentNotFound_JSONMarshalNotAttempted() {
	this.client.response = NotFoundResponse
	this.read()
	this.So(this.document.ID, should.Equal, 0)
	this.So(this.stdout.String(), should.ContainSubstring, "Document not found at '/document/path'\n")
}

func (this *DocumentReaderFixture) TestNilResponseBody() {
	this.client.response = BodyNilResponse
	this.assertPanic("HTTP response body was nil")
}

func (this *DocumentReaderFixture) TestBodyUnreadable() {
	this.client.response = BodyReadErrorResponse
	this.So(this.read, should.Panic)
	this.So(this.document.ID, should.Equal, 0)
	this.So(BodyReadErrorResponse.Body.(*FakeHTTPResponseBody).closed, should.BeTrue)
}

func (this *DocumentReaderFixture) TestBadJSON() {
	this.client.response = BadJSONResponse
	this.So(this.read, should.Panic)
	this.So(this.document.ID, should.Equal, 0)
	this.So(BodyReadErrorResponse.Body.(*FakeHTTPResponseBody).closed, should.BeTrue)
}

func (this *DocumentReaderFixture) TestValidUncompressedResponse_PopulatesDocument() {
	this.client.response = ValidUncompressedResponse
	this.read()
	this.So(this.document.ID, should.Equal, 1234)
	this.So(this.stdout.String(), should.BeEmpty)
	this.So(BodyReadErrorResponse.Body.(*FakeHTTPResponseBody).closed, should.BeTrue)
}
func (this *DocumentReaderFixture) TestValidCompressedResponse_PopulatesDocument() {
	this.client.response = ValidCompressedResponse
	this.read()
	this.So(this.document.ID, should.Equal, 1234)
	this.So(this.stdout.String(), should.BeEmpty)
	this.So(BodyReadErrorResponse.Body.(*FakeHTTPResponseBody).closed, should.BeTrue)
}
func (this *DocumentReaderFixture) read() {
	this.reader.ReadPanic(this.path, this.document)
}
func (this *DocumentReaderFixture) assertPanic(message string) {
	this.So(this.read, should.Panic)
	this.So(this.document.ID, should.Equal, 0)
}

//////////////////////////////////////////////////////////////////////////////////////////////

type FakeHTTPGetClient struct {
	err      error
	response *http.Response
	called   bool
}

func (this *FakeHTTPGetClient) Do(request *http.Request) (*http.Response, error) {
	this.called = true
	return this.response, this.err
}

///////////////////////////////////////////////////////////////////////////////////////////

type Document struct{ ID int }

////////////////////////////////////////////////////////////////////////////////////////////

var (
	NotFoundResponse          = &http.Response{StatusCode: 404, Body: newHTTPBody("Not found")}
	BodyNilResponse           = &http.Response{StatusCode: 200, Body: nil}
	BodyReadErrorResponse     = &http.Response{StatusCode: 200, Body: newReadErrorHTTPBody()}
	BadJSONResponse           = &http.Response{StatusCode: 200, Body: newHTTPBody("I am bad JSON.")}
	ValidCompressedResponse   = &http.Response{StatusCode: 200, Body: newHTTPBody(`{"ID": 1234}`)}
	ValidUncompressedResponse = &http.Response{StatusCode: 200, Body: newHTTPBody(`{"ID": 1234}`)}
)

func init() {
	ValidCompressedResponse.Header = make(http.Header)
	ValidCompressedResponse.Header.Set("Content-Encoding", "gzip")

	targetBuffer := bytes.NewBuffer([]byte{})
	writer := gzip.NewWriter(targetBuffer)
	io.Copy(writer, ValidCompressedResponse.Body)
	writer.Close()

	ValidCompressedResponse.Body = ioutil.NopCloser(targetBuffer)
}

func newHTTPBody(message string) io.ReadCloser {
	return &FakeHTTPResponseBody{Reader: strings.NewReader(message)}
}
func newReadErrorHTTPBody() io.ReadCloser {
	return &FakeHTTPResponseBody{err: errors.New("BOINK!")}
}

type FakeHTTPResponseBody struct {
	*strings.Reader

	err    error
	closed bool
}

func (this *FakeHTTPResponseBody) Read(p []byte) (int, error) {
	if this.err != nil {
		return 0, this.err
	}
	return this.Reader.Read(p)
}

func (this *FakeHTTPResponseBody) Close() error {
	this.closed = true
	return nil
}