// Copyright 2022 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package service_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/moov-io/bai2/pkg/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	testFileName = "sample1.txt"
	testDetailsWithNewlineTermination = "sample4-adhoc-continuations.txt"
	testDetailsWithSlashTermination = "sample4-with-terminators.txt"
	testDetailsWithInvalidCharacters = "sample5-invalid-characters-in-text.txt"
)

type HandlersTest struct {
	suite.Suite
	testServer *mux.Router
}

func (suite *HandlersTest) makeRequest(method, url, body string) (*httptest.ResponseRecorder, *http.Request) {
	request, err := http.NewRequest(method, url, strings.NewReader(body))
	assert.Equal(suite.T(), nil, err)
	recorder := httptest.NewRecorder()
	return recorder, request
}

func (suite *HandlersTest) getWriter(name string) (*multipart.Writer, *bytes.Buffer) {

	path := filepath.Join("..", "..", "test", "testdata", name)
	file, err := os.Open(path)
	assert.Equal(suite.T(), nil, err)

	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("input", filepath.Base(path))
	assert.Equal(suite.T(), nil, err)

	_, err = io.Copy(part, file)
	assert.Equal(suite.T(), nil, err)
	return writer, body
}

func (suite *HandlersTest) SetupTest() {

	suite.testServer = mux.NewRouter()

	err := service.ConfigureHandlers(suite.testServer)
	assert.Equal(suite.T(), nil, err)
}

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTest))
}

func (suite *HandlersTest) TestUnknownRequest() {
	recorder, request := suite.makeRequest(http.MethodGet, "/unknown", "")
	suite.testServer.ServeHTTP(recorder, request)
	assert.Equal(suite.T(), http.StatusNotFound, recorder.Code)
}

func (suite *HandlersTest) TestHealth() {
	recorder, request := suite.makeRequest(http.MethodGet, "/health", "")
	suite.testServer.ServeHTTP(recorder, request)
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)
}

func (suite *HandlersTest) TestPrint() {

	writer, body := suite.getWriter(testFileName)

	err := writer.Close()
	assert.Equal(suite.T(), nil, err)

	recorder, request := suite.makeRequest(http.MethodPost, "/print", body.String())
	request.Header.Set("Content-Type", writer.FormDataContentType())

	suite.testServer.ServeHTTP(recorder, request)
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)

	// Verify that the printed file matches the input file.
	path := filepath.Join("..", "..", "test", "testdata", testFileName)
	fixture, err := os.ReadFile(path)
	assert.Equal(suite.T(), nil, err)

	// NB. Account continuations are currently not written to file exactly as they were read.
	// Because of this behavior, the returned body does NOT strictly match the file data.
	// This test currently asserts on the shape of the file as created by the current return.
	// The difference between this output and the sample file is that a subset of data provided on
	// each account continuation (88) is instead output on the Account record (03).
	assert.NotEqual(suite.T(), recorder.Body.String(), string(fixture))

	expectedFileBody := `01,0004,12345,060321,0829,001,80,1,2/
02,12345,0004,1,060317,,CAD,/
03,10200123456,CAD,040,+000000000000,,,045,+000000000000,,,100,000000000208500/
88,3,V,060316,,400,000000000208500,8,V,060316,/
16,409,000000000002500,V,060316,,,,RETURNED CHEQUE     /
16,409,000000000090000,V,060316,,,,RTN-UNKNOWN         /
16,409,000000000000500,V,060316,,,,RTD CHQ SERVICE CHRG/
16,108,000000000203500,V,060316,,,,TFR 1020 0345678    /
16,108,000000000002500,V,060316,,,,MACLEOD MALL        /
16,108,000000000002500,V,060316,,,,MASCOUCHE QUE       /
16,409,000000000020000,V,060316,,,,1000 ISLANDS MALL   /
16,409,000000000090000,V,060316,,,,PENHORA MALL        /
16,409,000000000002000,V,060316,,,,CAPILANO MALL       /
16,409,000000000002500,V,060316,,,,GALERIES LA CAPITALE/
16,409,000000000001000,V,060316,,,,PLAZA ROCK FOREST   /
49,+00000000000834000,14/
03,10200123456,CAD,040,+000000000000,,,045,+000000000000,,,100,000000000111500/
88,2,V,060317,,400,000000000111500,4,V,060317,/
16,108,000000000011500,V,060317,,,,TFR 1020 0345678    /
16,108,000000000100000,V,060317,,,,MONTREAL            /
16,409,000000000100000,V,060317,,,,GRANDFALL NB        /
16,409,000000000009000,V,060317,,,,HAMILTON ON         /
16,409,000000000002000,V,060317,,,,WOODSTOCK NB        /
16,409,000000000000500,V,060317,,,,GALERIES RICHELIEU  /
49,+00000000000446000,9/
98,+00000000001280000,2,25/
99,+00000000001280000,1,27/`
	assert.Equal(suite.T(), recorder.Body.String(), expectedFileBody)
}

func (suite *HandlersTest) TestParse() {

	writer, body := suite.getWriter(testFileName)
	err := writer.Close()
	assert.Equal(suite.T(), nil, err)

	recorder, request := suite.makeRequest(http.MethodPost, "/parse", body.String())
	request.Header.Set("Content-Type", writer.FormDataContentType())

	suite.testServer.ServeHTTP(recorder, request)
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)
}

func (suite *HandlersTest) TestParse_Bai2FileWithInvalidCharacters() {

	writer, body := suite.getWriter(testDetailsWithInvalidCharacters)
	err := writer.Close()
	assert.Equal(suite.T(), nil, err)

	recorder, request := suite.makeRequest(http.MethodPost, "/parse", body.String())
	request.Header.Set("Content-Type", writer.FormDataContentType())

	suite.testServer.ServeHTTP(recorder, request)
	assert.Equal(suite.T(), http.StatusBadRequest, recorder.Code)
}

func (suite *HandlersTest) TestPrint_Bai2FileWithAdhocDetails() {

	writer, body := suite.getWriter(testDetailsWithNewlineTermination)
	err := writer.Close()
	assert.Equal(suite.T(), nil, err)

	recorder, request := suite.makeRequest(http.MethodPost, "/print", body.String())
	request.Header.Set("Content-Type", writer.FormDataContentType())

	suite.testServer.ServeHTTP(recorder, request)
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)

	expectedFileBody := `01,GSBI,cont001,210706,1249,1,,,2/
02,cont001,026015079,1,230906,2000,,/
03,107049924,USD,,,,,060,13053325440,,,100,000,,,400,000,,/
49,13053325440,2/
03,107049932,USD,,,,,060,6865898,,,100,1912,1,,400,000,,/
16,447,60000,,SPB2322984714570,1111,ACH Credit Payment/
16,557,200000,,SB2322600000214,021000080000030,ACH Credit Receipt Return/
49,000,2/
03,280000010657,USD,,,,,060,000,,,100,000,,,400,000,,/
49,000,2/
98,13060195162,4,16/
99,13060195162,1,18/`
	assert.Equal(suite.T(), recorder.Body.String(), expectedFileBody)
}

// This test and the test above use the same file -- the only difference is, one file does not include slash terminators
// for Detail records with Text continuations. Those records are instead newline terminated. When manually modifying this
// file to include Terminators, we parse out 17 Detail records rather than 2.
func (suite *HandlersTest) TestPrint_Bai2FileWithAdhocDetails_WithTerminators() {

	writer, body := suite.getWriter(testDetailsWithSlashTermination)
	err := writer.Close()
	assert.Equal(suite.T(), nil, err)

	recorder, request := suite.makeRequest(http.MethodPost, "/print", body.String())
	request.Header.Set("Content-Type", writer.FormDataContentType())

	suite.testServer.ServeHTTP(recorder, request)
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)

	expectedFileBody := `01,GSBI,cont001,210706,1249,1,,,2/
02,cont001,026015079,1,230906,2000,,/
03,107049924,USD,,,,,060,13053325440,,,100,000,,,400,000,,/
49,13053325440,2/
03,107049932,USD,,,,,060,6865898,,,100,1912,1,,400,000,,/
16,447,60000,,SPB2322984714570,1111,ACH Credit Payment/
16,261,143500,,SB2322600000404,GSQ4FBGFDGWGKY,ACH Credit Reject/
16,447,928650,,SPB2322684598521,AB-GS-RPFILERP0001-RPBA0001,ACH Credit Payment/
49,-1260161341762,26/
03,104108339,USD,010,159581194,,,015,159381194,,,040,158568897,,,045,158368897,,,100,000,,,400,200000,1,/
16,557,200000,,SB2322600000214,021000080000030,ACH Credit Receipt Return/
16,451,55555,,SB2322600000455,021000020000021,ACH Debit Payment/
16,266,1912,,GI2118700002010,20210706MMQFMPU8000001,Outgoing Wire Return/
16,495,50500,,GI2321400000090,GSV0DL6RKT,Outgoing Wire/
16,195,1125,,GI2229300000187,GS0D9VGMP1IWPLW,Incoming Wire/
16,257,60000,,SB2225800001203,028000020000335,ACH Debit Payment Return/
16,255,931,,SC2134800001999,,Check Return/
16,195,50050,,GI2228400005800,RTR60880840833,RTP Incoming/
16,175,527,,SX22293073766088,GS4N04L1COP45VY,Check Deposit/
16,475,10100,,SC2229300000152,01030340329,Check Paid/
16,275,337686,,GI2318000014342,e457328416d411eeaf020a58a9feac02,Cash Concentration/
16,165,5000,,SPB2321284264201,AB-GS-DDFILEAB0001-DDBAB0001,ACH Debit Collection/
16,475,44250,,SC2323300002416,8ce1829175a74ec88d67010dd7fb6132,Check Paid/
16,495,30000000,,GI2323300009168,3785726,Outgoing Wire/
49,6869722,8/
03,260000033037,USD,,,,,060,000,,,100,000,,,400,000,,/
49,000,2/
03,280000010657,USD,,,,,060,000,,,100,000,,,400,000,,/
49,000,2/
98,13060195162,4,16/
99,13060195162,1,18/`
	assert.Equal(suite.T(), recorder.Body.String(), expectedFileBody)
}
