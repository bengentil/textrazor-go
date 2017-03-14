package textrazor

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

const (
	testAPIKey = "1234567890"
	testURL    = "https://news.google.com"
	testText   = "Barclays misled shareholders and the public about one of the biggest investments in the bank's history, a BBC Panorama investigation has found."

	successful = true
	failed     = false

	errorResponseBody = `{
	    "ok": false,
	    "response": {
	    }
	}`
)

func checkHTTPResponse(t *testing.T, r *HTTPResponse) {
	if r == nil {
		t.Error("expect HTTPResponse, got nil")
		t.FailNow()
	}
	if r.Status != http.StatusOK {
		t.Error("expect 'HTTPStatus' field to be http.StatusOK (http.StatusOK), got", r.Status)
	}
	if r.Headers.Get("Content-Type") != "application/json" {
		t.Error("expect 'Content-Type' header in response to be application/json, got", r.Headers.Get("Content-Type"))
	}
	if !r.Ok {
		t.Error("expect 'ok' field in response to be true, got", r.Ok)
	}
}

//***************************************************************
// 			fakeTransport
// minimal http.RoundTripper implementation
// to avoid making real API call during tests
type fakeTransport struct {
	t              *testing.T
	responseStatus int
	responseBody   string
	shouldFail     bool
}

func FakeTransport(t *testing.T, responseStatus int, responseBody string, shouldFail bool) http.RoundTripper {
	return &fakeTransport{t: t, responseStatus: responseStatus, responseBody: responseBody, shouldFail: shouldFail}
}

func (t *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.shouldFail {
		return nil, fmt.Errorf("expected error")
	}

	headers := ""
	if len(req.Header) > 0 {
		for k, v := range req.Header {
			headers = "-H '" + k + ": " + v[0] + "'"
		}
	}
	body, _ := ioutil.ReadAll(req.Body)
	t.t.Log("curl -X", req.Method, "-d '"+string(body)+"'", headers, req.URL)

	response := &http.Response{
		Header:     make(http.Header),
		Request:    req,
		StatusCode: t.responseStatus,
	}
	response.Header.Set("Content-Type", "application/json")

	if t.responseBody == "FAKE_READ_ISSUE" {
		response.Body = &faultyReader{}
	} else {
		response.Body = ioutil.NopCloser(strings.NewReader(t.responseBody))
	}

	return response, nil
}

//***************************************************************
// 			faultyReader
// minimal io.ReadCloser implementation
// to simulate unexpected error while reading http response
type faultyReader struct{}

func (r *faultyReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("expected error")
}

func (r *faultyReader) Close() (err error) { return nil }

//***************************************************************
// 			Analyze, AnalyzeText, AnalyzeURL tests
const analyseResponseBody = `{
    "response": {
        "sentences": [
            {
                "position": 0,
                "words": [
                    {
                        "position": 0,
                        "startingPos": 0,
                        "endingPos": 3,
                        "stem": "bbc",
                        "lemma": "bbc",
                        "token": "BBC",
                        "partOfSpeech": "NNP"
                    },
                    {
                        "position": 1,
                        "startingPos": 3,
                        "endingPos": 3,
                        "stem": ".",
                        "lemma": ".",
                        "token": ".",
                        "partOfSpeech": "."
                    }
                ]
            }
        ],
        "language": "eng",
        "languageIsReliable": true,
        "entities": [
            {
                "id": 0,
                "type": [
                    "Agent",
                    "Organisation",
                    "Company",
                    "Broadcaster",
                    "TelevisionStation"
                ],
                "matchingTokens": [
                    0
                ],
                "entityId": "BBC",
                "freebaseTypes": [
                    "/film/film_distributor",
                    "/tv/tv_network",
                    "/business/customer",
                    "/award/award_presenting_organization",
                    "/book/book_subject",
                    "/media_common/netflix_genre",
                    "/organization/organization_founder",
                    "/business/employer",
                    "/broadcast/radio_station_owner",
                    "/computer/software_developer",
                    "/film/production_company",
                    "/business/consumer_company",
                    "/award/award_nominee",
                    "/broadcast/tv_station_owner",
                    "/award/award_winner",
                    "/book/author",
                    "/book/periodical_publisher",
                    "/tv/tv_program_creator",
                    "/radio/radio_subject",
                    "/organization/organization",
                    "/broadcast/artist",
                    "/internet/website_owner",
                    "/broadcast/producer",
                    "/tv/tv_producer",
                    "/business/business_operation"
                ],
                "confidenceScore": 1.726,
                "wikiLink": "http://en.wikipedia.org/wiki/BBC",
                "matchedText": "BBC",
                "freebaseId": "/m/0ncl8zk",
                "relevanceScore": 0,
                "entityEnglishId": "BBC",
                "startingPos": 0,
                "endingPos": 3,
                "wikidataId": "Q9531"
            }
        ]
    },
    "time": 0.003359,
    "ok": true
}`

var analyzeTests = []struct {
	expectedResult                bool
	responseStatus                int
	responseBody                  string
	useCompression, useEncryption bool
	endpoint, secureEndpoint      string
	params                        Params
}{
	{successful, http.StatusOK, analyseResponseBody, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, Params{"text": {testText}, "extractors": {"entities", "entailments"}}},
	{successful, http.StatusOK, analyseResponseBody, false, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, Params{"text": {testText}, "extractors": {"entities", "entailments"}}},
	{successful, http.StatusOK, analyseResponseBody, true, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, Params{"text": {testText}, "extractors": {"entities", "entailments"}}},
	{successful, http.StatusOK, analyseResponseBody, DefaultUseCompression, false, DefaultEndpoint, DefaultSecureEndpoint, Params{"text": {testText}, "extractors": {"entities", "entailments"}}},
	{successful, http.StatusOK, analyseResponseBody, DefaultUseCompression, true, DefaultEndpoint, DefaultSecureEndpoint, Params{"text": {testText}, "extractors": {"entities", "entailments"}}},
	{failed, http.StatusOK, analyseResponseBody, DefaultUseCompression, false, "INVALID_URL!!!", DefaultSecureEndpoint, Params{"text": {testText}, "extractors": {"entities", "entailments"}}},
	{failed, http.StatusOK, analyseResponseBody, DefaultUseCompression, true, DefaultEndpoint, "INVALID_URL!!!", Params{"text": {testText}, "extractors": {"entities", "entailments"}}},
	{failed, http.StatusOK, analyseResponseBody, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, Params{"extractors": {"entities", "entailments"}}},
	{successful, http.StatusOK, analyseResponseBody, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, Params{"url": {testURL}, "extractors": {"entities", "entailments"}}},
	{failed, http.StatusOK, analyseResponseBody, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, Params{"text": {testText}, "url": {testURL}, "extractors": {"entities", "entailments"}}},
	{failed, http.StatusOK, analyseResponseBody, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, Params{"text": {testText}}},
	{failed, http.StatusOK, analyseResponseBody, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, nil},
	{failed, http.StatusServiceUnavailable, analyseResponseBody, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, Params{"text": {testText}, "extractors": {"entities", "entailments"}}},
	{failed, http.StatusOK, errorResponseBody, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, Params{"text": {testText}, "extractors": {"entities", "entailments"}}},
}

func TestAnalyze(t *testing.T) {
	for i, tst := range analyzeTests {
		t.Log("TestAnalyze[", i, "]")
		client := NewCustomClient(testAPIKey, tst.useCompression, tst.useEncryption, tst.endpoint, tst.secureEndpoint, FakeTransport(t, tst.responseStatus, tst.responseBody, false))
		analysis, err := client.Analyze(tst.params)
		if err != nil {
			t.Log(err)
			if tst.expectedResult == successful {
				t.Error(err)
			}
		} else {
			if tst.expectedResult == failed {
				t.Error("this test should fail:", tst)
			}

			checkHTTPResponse(t, analysis.HTTPResponse)

			if len(analysis.Entities) != 1 && analysis.Entities[0].EntityID != "BBC" {
				if len(analysis.Entities) > 0 {
					t.Error("expect 1 entity in response with EntityID=='BBC', got", len(analysis.Entities), "entities, first one EntityID is", analysis.Entities[0].EntityID)
				} else {
					t.Error("expect 1 entity in response with EntityID=='BBC', got", len(analysis.Entities), "entities")
				}
			}
		}
	}
}

func TestAnalyzeText(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, analyseResponseBody, false))
	analysis, err := client.AnalyzeText(testText, Params{"extractors": {"entities", "entailments"}})
	if err != nil {
		t.Error(err)
	}
	checkHTTPResponse(t, analysis.HTTPResponse)
}

func TestAnalyzeURL(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, analyseResponseBody, false))
	analysis, err := client.AnalyzeURL(testURL, Params{"extractors": {"entities", "entailments"}})
	if err != nil {
		t.Error(err)
	}
	checkHTTPResponse(t, analysis.HTTPResponse)
}

//***************************************************************
// 			Account tests

const accountResponseBody = `{
    "ok": true,
    "response": {
        "requestsUsedToday": 17,
        "concurrentRequestsUsed": 0,
        "concurrentRequestLimit": 2,
        "plan": "FREE",
        "planDailyRequestsIncluded": 500
    }
}`

func TestGetAccount(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, accountResponseBody, false))
	account, err := client.GetAccount()
	if err != nil {
		t.Error(err)
	}
	checkHTTPResponse(t, account.HTTPResponse)

	if account.Plan != "FREE" {
		t.Error("expected account.Plan == FREE, got", account.Plan)
	}
	if account.RequestsUsedToday != 17 {
		t.Error("expected account.RequestsUsedToday == 17, got", account.RequestsUsedToday)
	}
	if account.ConcurrentRequestLimit != 2 {
		t.Error("expected account.ConcurrentRequestLimit == 2, got", account.ConcurrentRequestLimit)
	}
	if account.PlanDailyIncludedRequests != 500 {
		t.Error("expected account.PlanDailyIncludedRequests == 500, got", account.PlanDailyIncludedRequests)
	}
}

func TestGetAccountError(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, errorResponseBody, false))
	_, err := client.GetAccount()
	if err != nil {
		t.Log(err)
	}
	if err == nil {
		t.Error("this test should fail")
	}
}

//***************************************************************
// 			Dictionary tests

const (
	dictCreateResponseBody  = `{"time":0.004913,"ok":true}`
	dictDeleteResponseBody  = `{"time":0.004913,"ok":true}`
	dictGetDictionariesBody = `{"dictionaries":[{"id":"test_ents","matchType":"TOKEN","caseInsensitive":true,"language":"eng"}],"time":0.002655,"ok":true}`
	dictGetDictBody         = `{"response":{"id":"test_ents","matchType":"TOKEN","caseInsensitive":true,"language":"eng"},"time":0.002503,"ok":true}`
	dictGetDictEntriesBody  = `{"response":{"offset":0,"limit":20,"total":1,"entries":[{"id":"DEV2","text":"Bjarne Stroustrup","data":{}}]},"time":0.005158,"ok":true}`
	dictGetDictEntryBody    = `{"response":{"id":"DEV2","text":"Bjarne Stroustrup","data":{}},"time":0.001278,"ok":true}`

	dictID        = "test_ents"
	dictEntryID   = "DEV2"
	dictEntryText = "Bjarne Stroustrup"
)

func TestCreateDictionary(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, dictCreateResponseBody, false))
	d := &Dictionary{MatchType: "token", CaseInsensitive: true, Language: "eng"}
	resp, err := client.CreateDictionary(d)
	if err != nil {
		t.Error(err)
	}
	checkHTTPResponse(t, resp)
}

func TestGetDictionaries(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, dictGetDictionariesBody, false))
	resp, err := client.GetDictionaries()
	if err != nil {
		t.Error(err)
	}
	checkHTTPResponse(t, resp)
	if len(resp.Dictionaries) != 1 && resp.Dictionaries[0].ID != dictID {
		if len(resp.Dictionaries) > 0 {
			t.Error("expect 1 dictionary in response with ID==", dictID, "got", len(resp.Dictionaries), "entities, first one EntityID is", resp.Dictionaries[0].ID)
		} else {
			t.Error("expect 1 dictionary in response with ID==", dictID, "got", len(resp.Dictionaries), "entities")
		}
	}
}

func TestGetDictionary(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, dictGetDictBody, false))
	dict, err := client.GetDictionary(dictID)
	if err != nil {
		t.Error(err)
	}
	checkHTTPResponse(t, dict.HTTPResponse)
	if dict.ID != dictID {
		t.Error("expect Dictionary ID==", dictID, "got", dict.ID)
	}
}

func TestGetDictionaryError(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, errorResponseBody, false))
	_, err := client.GetDictionary(dictID)
	if err != nil {
		t.Log(err)
	}
	if err == nil {
		t.Error("this test should fail")
	}
}

func TestDeleteDictionary(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, dictDeleteResponseBody, false))
	resp, err := client.DeleteDictionary(dictID)
	if err != nil {
		t.Error(err)
	}
	checkHTTPResponse(t, resp)
}

func TestAddDictionaryEntry(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, dictCreateResponseBody, false))
	e := &DictionaryEntry{ID: dictEntryID, Text: dictEntryText}
	resp, err := client.AddDictionaryEntry(dictID, e)
	if err != nil {
		t.Error(err)
	}
	checkHTTPResponse(t, resp)
}

func TestGetDictionaryEntries(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, dictGetDictEntriesBody, false))
	el, err := client.GetDictionaryEntries(dictID, 20, 0)
	if err != nil {
		t.Error(err)
	}
	checkHTTPResponse(t, el.HTTPResponse)
	if len(el.Entries) != 1 && el.Entries[0].ID != dictEntryID && el.Entries[0].Text != dictEntryText {
		if len(el.Entries) > 0 {
			t.Error("expect 1 dictionary entry in response with ID==", dictID, "got", len(el.Entries), "entities, first one ID is", el.Entries[0].ID)
		} else {
			t.Error("expect 1 dictionary entry in response with ID==", dictID, "got", len(el.Entries), "entities")
		}
	}
}

func TestGetDictionaryEntriesError(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, errorResponseBody, false))
	_, err := client.GetDictionaryEntries(dictID, 20, 0)
	if err != nil {
		t.Log(err)
	}
	if err == nil {
		t.Error("this test should fail")
	}
}

func TestGetDictionaryEntry(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, dictGetDictEntryBody, false))
	e, err := client.GetDictionaryEntry(dictID, dictEntryID)
	if err != nil {
		t.Error(err)
	}
	checkHTTPResponse(t, e.HTTPResponse)
	if e.ID != dictEntryID {
		t.Error("expect Entry ID==", dictEntryID, "got", e.ID)
	}
	if e.Text != dictEntryText {
		t.Error("expect Entry Text==", dictEntryText, "got", e.Text)
	}
}

func TestGetDictionaryEntryError(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, errorResponseBody, false))
	_, err := client.GetDictionaryEntry(dictID, dictEntryID)
	if err != nil {
		t.Log(err)
	}
	if err == nil {
		t.Error("this test should fail")
	}
}

func TestDeleteDictionaryEntry(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, dictDeleteResponseBody, false))
	resp, err := client.DeleteDictionaryEntry(dictID, dictEntryID)
	if err != nil {
		t.Error(err)
	}
	checkHTTPResponse(t, resp)
}

//***************************************************************
// 			Category tests
const (
	catCreateResponseBody        = `{"time":0.007047,"ok":true}`
	catGetCategoriesResponseBody = `{
    "response": {
        "id": "sport2",
        "offset": 0,
        "limit": 20,
        "total": 3,
        "lastUpdated": 1489517818,
        "categories": [
            {
                "categoryId": "100",
                "label": "Golf",
                "query": "concept('sport>golf')"
            },
            {
                "categoryId": "101",
                "label": "Squash",
                "query": "concept('sport>squash')"
            },
            {
                "categoryId": "102",
                "label": "Cricket",
                "query": "concept('sport>cricket')"
            }
        ]
    },
    "time": 0.00612,
    "ok": true
}`
	catGetResponseBody    = `{"response":{"categoryId":"100","label":"Golf","query":"concept('sport>golf')"},"time":0.00153,"ok":true}`
	catDeleteResponseBody = `{"time":0.003754,"ok":true}`

	catJSON = `[
{"categoryId":"100","label":"Golf","query":"concept('sport>golf')"},
{"categoryId":"101","label":"Squash","query":"concept('sport>squash')"},
{"categoryId":"102","label":"Cricket","query":"concept('sport>cricket')"}
]
`
	catCSV = `100,Golf,concept('sport>golf')
101,Squash,concept('sport>squash')
102,Cricket,concept('sport>cricket')
`

	catDictID = "sport"
	catID     = "100"
	catLabel  = "Golf"
	catQuery  = "concept('sport>golf')"
)

func TestCreateClassifierFromJSON(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, catCreateResponseBody, false))
	resp, err := client.CreateClassifierFromJSON(catDictID, catJSON)
	if err != nil {
		t.Error(err)
	}
	checkHTTPResponse(t, resp)
}

func TestCreateClassifierFromCSV(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, catCreateResponseBody, false))
	resp, err := client.CreateClassifierFromCSV(catDictID, catCSV)
	if err != nil {
		t.Error(err)
	}
	checkHTTPResponse(t, resp)
}

func TestDeleteClassifier(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, catDeleteResponseBody, false))
	resp, err := client.DeleteClassifier(catDictID)
	if err != nil {
		t.Error(err)
	}
	checkHTTPResponse(t, resp)
}

func TestGetClassifierCategories(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, catGetCategoriesResponseBody, false))
	cat, err := client.GetClassifierCategories(catDictID, 20, 0)
	if err != nil {
		t.Error(err)
	}
	checkHTTPResponse(t, cat.HTTPResponse)
	if len(cat.Categories) != 3 && cat.Categories[0].CategoryID != catID && cat.Categories[0].Label != catLabel {
		if len(cat.Categories) > 0 {
			t.Error("expect 1 dictionary entry in response with ID==", catID, "got", len(cat.Categories), "entities, first one ID is", cat.Categories[0].CategoryID)
		} else {
			t.Error("expect 1 dictionary entry in response with ID==", catID, "got", len(cat.Categories), "entities")
		}
	}
}

func TestGetClassifierCategoriesError(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, errorResponseBody, false))
	_, err := client.GetClassifierCategories(catDictID, 20, 0)
	if err != nil {
		t.Log(err)
	}
	if err == nil {
		t.Error("this test should fail")
	}
}

func TestGetClassifierCategory(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, catGetResponseBody, false))
	cat, err := client.GetClassifierCategory(catDictID, catID)
	if err != nil {
		t.Error(err)
	}
	checkHTTPResponse(t, cat.HTTPResponse)
	if cat.CategoryID != catID {
		t.Error("expect Entry ID==", dictEntryID, "got", cat.CategoryID)
	}
	if cat.Label != catLabel {
		t.Error("expect Entry Label==", catLabel, "got", cat.Label)
	}
	if cat.Query != catQuery {
		t.Error("expect Entry Query==", catQuery, "got", cat.Query)
	}
}

func TestGetClassifierCategoryError(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, errorResponseBody, false))
	_, err := client.GetClassifierCategory(catDictID, catID)
	if err != nil {
		t.Log(err)
	}
	if err == nil {
		t.Error("this test should fail")
	}
}

func TestDeleteClassifierCategory(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, catDeleteResponseBody, false))
	resp, err := client.DeleteClassifierCategory(catDictID, catID)
	if err != nil {
		t.Error(err)
	}
	checkHTTPResponse(t, resp)
}

//***************************************************************
// 			HTTP error handling tests
func TestTransportFailure(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, "", true))
	_, err := client.Analyze(Params{"text": {testText}, "extractors": {"entities", "entailments"}})
	if err != nil {
		t.Log(err)
	}
	if err == nil {
		t.Error("this test should fail")
	}
}

func TestHTTPRequestFailure(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, "", false))
	_, err := client.doRequest("/", "INVALID_METHOD€€€", nil, nil, &Analysis{})
	if err != nil {
		t.Log(err)
	}
	if err == nil {
		t.Error("this test should fail")
	}
}

func TestHTTPResponseFailure(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, "FAKE_READ_ISSUE", false))
	_, err := client.doRequest("/", http.MethodPost, nil, nil, &Analysis{})
	if err != nil {
		t.Log(err)
	}
	if err == nil {
		t.Error("this test should fail")
	}
}

func TestEmptyHTTPResponseBody(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, "", false))
	_, err := client.doRequest("/", http.MethodPost, nil, nil, &Analysis{})
	if err != nil {
		t.Log(err)
	}
	if err == nil {
		t.Error("this test should fail")
	}
}

type faultyBody struct{}

func (f *faultyBody) Encode() (string, error) { return "", fmt.Errorf("bad body") }

func TestHTTPRequestBodyFailure(t *testing.T) {
	client := NewCustomClient(testAPIKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, FakeTransport(t, http.StatusOK, "", false))
	_, err := client.doRequest("/", http.MethodPost, nil, &faultyBody{}, &Analysis{})
	if err != nil {
		t.Log(err)
	}
	if err == nil {
		t.Error("this test should fail")
	}
}

//***************************************************************
// 			Params tests
func TestParams(t *testing.T) {
	p := Params{"extractors": {"entities", "entailments"}}
	if p.Get("extractors") != "entities" {
		t.Error("p.Get should return the first value associated to the key")
	}

	p.Add("text", "text1")
	p.Add("text", "text2")
	if p.Get("text") != "text1" || len(p["text"]) != 2 || p["text"][1] != "text2" {
		t.Error("p.Add should add values to the key")
	}

	p.Set("text", testText)
	if p.Get("text") != testText {
		t.Error("p.Set should replace any existing values associated to the key")
	}

	p.Del("text")
	if p.Get("text") != "" {
		t.Error("p.Del should delete any existing values associated to the key")
	}

	pStr, _ := p.Encode()
	if pStr != "extractors=entities&extractors=entailments" {
		t.Error("p.Encode should encore in URL format")
	}
}
