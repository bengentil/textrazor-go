// Package textrazor implements a client for the TextRazor Text Analytics API.
//
// An API key is required and can be optained on https://www.textrazor.com
package textrazor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// default values used by NewDefaultClient
const (
	DefaultEndpoint       = "http://api.textrazor.com"
	DefaultSecureEndpoint = "https://api.textrazor.com"
	DefaultUseCompression = true
	DefaultUseEncryption  = true
)

const (
	// HTTP header used for Authentication
	apiKeyHeader = "X-TextRazor-Key"
	// Content-Types
	contentTypeJSON = "application/json"
	contentTypeCSV  = "application/csv"
	contentTypeURL  = "application/x-www-form-urlencoded"
)

// Params defines a generic map for endpoints parameters
type Params url.Values

// Add is similar to https://golang.org/pkg/net/url/#Values.Add
func (p Params) Add(key, value string) {
	url.Values(p).Add(key, value)
}

// Del is similar to https://golang.org/pkg/net/url/#Values.Del
func (p Params) Del(key string) {
	url.Values(p).Del(key)
}

// Encode is similar to https://golang.org/pkg/net/url/#Values.Encode
//
// returns a nil error to comply with RequestBody interface
func (p Params) Encode() (string, error) {
	return url.Values(p).Encode(), nil
}

// Get is similar to https://golang.org/pkg/net/url/#Values.Get
func (p Params) Get(key string) string {
	return url.Values(p).Get(key)
}

// Set is similar to https://golang.org/pkg/net/url/#Values.Set
func (p Params) Set(key, value string) {
	url.Values(p).Set(key, value)
}

// RequestBody defines a interface to generate a request body from multiple structs
type RequestBody interface {
	Encode() (string, error)
}

type rawRequest struct {
	Body string
}

// Encode allows rawRequest to be compliant with RequestBody interface
func (r *rawRequest) Encode() (string, error) { return r.Body, nil }

// Response defines an interface for the response field in multiple API calls
type Response interface {
	setHTTPResponse(*HTTPResponse)
}

// EmptyResponse defines an empty struct to be able to parse the JSON when 'response' field doesn't exist in the HTTP response
type EmptyResponse struct{}

func (r *EmptyResponse) setHTTPResponse(*HTTPResponse) {}

// HTTPResponse https://www.textrazor.com/docs/rest#TextRazorResponse
type HTTPResponse struct {
	Status   int         `json:"-"`
	Headers  http.Header `json:"-"`
	Body     []byte      `json:"-"`
	Time     float32     `json:"time"`
	Response Response    `json:"response"`

	// FIXME: most replies returns an object called 'response', except for 'GET /entities/'
	// which returns a json array called 'dictionaries'
	Dictionaries []Dictionary `json:"dictionaries"`

	Ok      bool   `json:"ok"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// ParseBody parses JSON HTTP Body
func (r *HTTPResponse) ParseBody() error {
	return json.Unmarshal(r.Body, r)
}

// Analysis https://www.textrazor.com/docs/rest#TextRazorResponse
type Analysis struct {
	HTTPResponse           *HTTPResponse    `json:"-"`
	CustomAnnotationOutput string           `json:"customAnnotationOutput"`
	CleanedText            string           `json:"cleanedText"`
	RawText                string           `json:"rawText"`
	Entailments            []Entailment     `json:"entailments"`
	Entities               []Entity         `json:"entities"`
	Topics                 []Topic          `json:"topics"`
	Categories             []ScoredCategory `json:"categories"`
	NounPhrases            []NounPhrase     `json:"nounPhrases"`
	Properties             []Property       `json:"properties"`
	Relations              []Relation       `json:"relations"`
	Sentences              []Sentence       `json:"sentences"`
	MatchingRules          []string         `json:"matchingRules"`
}

func (a *Analysis) setHTTPResponse(r *HTTPResponse) { a.HTTPResponse = r }

// Entity https://www.textrazor.com/docs/rest#Entity
type Entity struct {
	ID              int               `json:"id"`
	EntityID        string            `json:"entityId"`
	EntityEnglishID string            `json:"entityEnglishId"`
	CustomEntityID  string            `json:"customEntityId"`
	ConfidenceScore float32           `json:"confidenceScore"`
	Types           []string          `json:"type"`
	FreebaseTypes   []string          `json:"freebaseTypes"`
	FreebaseID      string            `json:"freebaseId"`
	WikidataID      string            `json:"wikidataId"`
	MatchingTokens  []int             `json:"matchingTokens"`
	MatchedText     string            `json:"matchedText"`
	Data            map[string]string `json:"data"`
	RelevanceScore  float32           `json:"relevanceScore"`
	WikiLink        string            `json:"wikiLink"`
}

// Topic https://www.textrazor.com/docs/rest#Topic
type Topic struct {
	Label      string  `json:"label"`
	Score      float32 `json:"score"`
	WikiLink   string  `json:"wikiLink"`
	WikidataID string  `json:"wikidataId"`
}

// ScoredCategory https://www.textrazor.com/docs/rest#ScoredCategory
type ScoredCategory struct {
	CategoryID   string  `json:"categoryId"`
	Label        string  `json:"label"`
	Score        float32 `json:"score"`
	ClassifierID string  `json:"classifierId"`
}

// Entailment https://www.textrazor.com/docs/rest#Entailment
type Entailment struct {
	ContextScore  float32           `json:"contextScore"`
	EntailedTree  map[string]string `json:"entailedTree"`
	WordPositions []int             `json:"wordPositions"`
	PriorScore    float32           `json:"priorScore"`
	Score         float32           `json:"score"`
}

// RelationType defines the type for the relation field in RelationParam
type RelationType string

// Valid options for relation field in RelationParam
const (
	SUBJECT RelationType = "SUBJECT"
	OBJECT  RelationType = "OBJECT"
	OTHER   RelationType = "OTHER"
)

// RelationParam https://www.textrazor.com/docs/rest#RelationParam
type RelationParam struct {
	WordPositions []int        `json:"wordPositions"`
	Relation      RelationType `json:"classifierId"`
}

// NounPhrase https://www.textrazor.com/docs/rest#NounPhrase
type NounPhrase struct {
	WordPositions []int `json:"wordPositions"`
}

// Property https://www.textrazor.com/docs/rest#Property
type Property struct {
	WordPositions     []int `json:"wordPositions"`
	PropertyPositions []int `json:"propertyPositions"`
}

// Relation https://www.textrazor.com/docs/rest#Relation
type Relation struct {
	Params        []RelationParam `json:"params"`
	WordPositions []int           `json:"wordPositions"`
}

// SenseScore defines a map with scores of each Wordnet sense the word may be a part of
type SenseScore map[string]float32

// SuggestionScore defines a map with scores of each spelling suggestion that might replace the word
type SuggestionScore map[string]float32

// Word https://www.textrazor.com/docs/rest#Word
type Word struct {
	EndingPos           int               `json:"endingPos"`
	StartingPos         int               `json:"startingPos"`
	Lemma               string            `json:"lemma"`
	ParentPosition      int               `json:"parentPosition"`
	PartOfSpeech        string            `json:"partOfSpeech"`
	Senses              []SenseScore      `json:"senses"`
	SpellingSuggestions []SuggestionScore `json:"spellingSuggestions"`
	Position            int               `json:"position"`
	RelationToParent    string            `json:"relationToParent"`
	Stem                string            `json:"stem"`
	Token               string            `json:"token"`
}

// Sentence https://www.textrazor.com/docs/rest#Sentence
type Sentence struct {
	Words []Word `json:"words"`
}

// Dictionary https://www.textrazor.com/docs/rest#Dictionary
type Dictionary struct {
	HTTPResponse    *HTTPResponse `json:"-"`
	MatchType       string        `json:"matchType"`
	CaseInsensitive bool          `json:"caseInsensitive"`
	ID              string        `json:"id"`
	Language        string        `json:"language"`
}

// Encode encodes the Dictionary struct in JSON
func (d *Dictionary) Encode() (string, error) {
	b, err := json.Marshal(d)
	return string(b), err
}

func (d *Dictionary) setHTTPResponse(r *HTTPResponse) { d.HTTPResponse = r }

// DictionaryEntry https://www.textrazor.com/docs/rest#DictionaryEntry
type DictionaryEntry struct {
	HTTPResponse *HTTPResponse     `json:"-"`
	ID           string            `json:"id"`
	Text         string            `json:"text"`
	Data         map[string]string `json:"data"`
}

func (e *DictionaryEntry) setHTTPResponse(r *HTTPResponse) { e.HTTPResponse = r }

// DictionaryEntryList defines the response for GetDictionaryEntries
type DictionaryEntryList struct {
	HTTPResponse *HTTPResponse     `json:"-"`
	Offset       int               `json:"offset"`
	Limit        int               `json:"limit"`
	Total        int               `json:"total"`
	Entries      []DictionaryEntry `json:"entries"`
}

// Encode encodes the DictionaryEntryList struct in JSON for AddDictionaryEntries
func (l *DictionaryEntryList) Encode() (string, error) {
	b, err := json.Marshal(l.Entries)
	return string(b), err
}

func (l *DictionaryEntryList) setHTTPResponse(r *HTTPResponse) { l.HTTPResponse = r }

// Category https://www.textrazor.com/docs/rest#Category
type Category struct {
	HTTPResponse *HTTPResponse `json:"-"`
	CategoryID   string        `json:"categoryId"`
	Label        string        `json:"label"`
	Query        string        `json:"query"`
}

func (c *Category) setHTTPResponse(r *HTTPResponse) { c.HTTPResponse = r }

// CategoryList response for GetClassifierCategory
type CategoryList struct {
	HTTPResponse *HTTPResponse `json:"-"`
	Offset       int           `json:"offset"`
	Limit        int           `json:"limit"`
	Total        int           `json:"total"`
	Categories   []Category    `json:"categories"`
}

func (l *CategoryList) setHTTPResponse(r *HTTPResponse) { l.HTTPResponse = r }

// Account https://www.textrazor.com/docs/rest#Account
type Account struct {
	HTTPResponse           *HTTPResponse `json:"-"`
	Plan                   string        `json:"plan"`
	ConcurrentRequestLimit int           `json:"concurrentRequestLimit"`
	ConcurrentRequestsUsed int           `json:"concurrentRequestsUsed"`
	// FIXME: the DOC says planDailyIncludedRequests but the api responds with planDailyRequestsIncluded
	PlanDailyIncludedRequests int `json:"planDailyRequestsIncluded"`
	RequestsUsedToday         int `json:"requestsUsedToday"`
}

func (a *Account) setHTTPResponse(r *HTTPResponse) { a.HTTPResponse = r }

// DefaultHeaders returns valid http.Header with Content-Type set
func DefaultHeaders(contentType string) http.Header {
	return http.Header{"Content-Type": {contentType}}
}

// DefaultTransport creates a compressed or uncompressed http.Transport
func DefaultTransport(useCompression bool) http.RoundTripper {
	return &http.Transport{
		DisableCompression: !useCompression,
	}
}

// Client defines a TextRazor http client
type Client struct {
	apiKey         string
	useCompression bool
	UseEncryption  bool
	Endpoint       string
	SecureEndpoint string
	httpTransport  http.RoundTripper
}

// NewClient returns a TextRazor client with default parameters
func NewClient(apiKey string) *Client {
	return NewCustomClient(apiKey, DefaultUseCompression, DefaultUseEncryption, DefaultEndpoint, DefaultSecureEndpoint, DefaultTransport(DefaultUseCompression))
}

// NewCustomClient returns a TextRazor client with custom parameters and custom transport
func NewCustomClient(apiKey string, useCompression, useEncryption bool, endpoint, secureEndpoint string, transport http.RoundTripper) *Client {
	return &Client{apiKey: apiKey,
		useCompression: useCompression,
		UseEncryption:  useEncryption,
		Endpoint:       endpoint,
		SecureEndpoint: secureEndpoint,
		httpTransport:  transport}
}

// doRequest execute a http request with the client parameters and transport
func (c *Client) doRequest(path, method string, headers http.Header, body RequestBody, response Response) (*HTTPResponse, error) {
	client := &http.Client{Transport: c.httpTransport}

	// set endpointURL
	endpointURL := c.Endpoint
	if c.UseEncryption {
		endpointURL = c.SecureEndpoint
	}

	// generate URL
	u, err := url.ParseRequestURI(endpointURL + path)
	if err != nil {
		return nil, fmt.Errorf("URI parsing failed '%v': %v", endpointURL+path, err)
	}

	// generate the request body
	bodyStr := ""
	if body != nil {
		bodyStr, err = body.Encode()
		if err != nil {
			return nil, fmt.Errorf("body request encoding failed: %v", err)
		}
	}

	// create a Request with the URL and the Body
	req, err := http.NewRequest(method, u.String(), bytes.NewBufferString(bodyStr))
	if err != nil {
		return nil, fmt.Errorf("http request creation failed: %v", err)
	}

	// set headers
	if headers != nil {
		req.Header = headers
	}
	req.Header.Add(apiKeyHeader, c.apiKey)

	// execute the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request execution failed: %v", err)
	}
	defer resp.Body.Close()

	// get the response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("http response body read failed: %v", err)
	}

	// build the response struct and decode json if request is successful
	httpResponse := &HTTPResponse{Status: resp.StatusCode, Headers: resp.Header, Body: respBody, Response: response}
	response.setHTTPResponse(httpResponse)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}
	err = httpResponse.ParseBody()
	if err != nil {
		return nil, fmt.Errorf("http response body parsing failed: %v", err)
	}

	if !httpResponse.Ok {
		return nil, fmt.Errorf("unexpected 'ok' field value: %v", httpResponse.Ok)
	}

	return httpResponse, nil
}

// Analyze returns a text analysis of either:
//
// * the text defined in the 'text' field
//
// * the web page refered by the 'url' field
//
// https://www.textrazor.com/docs/rest#analysis
func (c *Client) Analyze(params Params) (*Analysis, error) {
	analysis := &Analysis{}
	if (params.Get("text") == "" && params.Get("url") == "") || (params.Get("text") != "" && params.Get("url") != "") {
		return nil, fmt.Errorf("either 'url' or 'text' should be specified, not both")
	}
	if params.Get("extractors") == "" {
		return nil, fmt.Errorf("at least one 'extractors' should be specified")
	}
	if _, err := c.doRequest("/", http.MethodPost, DefaultHeaders(contentTypeURL), params, analysis); err != nil {
		return nil, err
	}
	return analysis, nil
}

// AnalyzeText returns a text analysis of the given text
func (c *Client) AnalyzeText(text string, params Params) (*Analysis, error) {
	params.Set("text", text)
	return c.Analyze(params)
}

// AnalyzeURL returns a text analysis of the given URL
func (c *Client) AnalyzeURL(urlStr string, params Params) (*Analysis, error) {
	params.Set("url", urlStr)
	return c.Analyze(params)
}

// GetAccount returns an Account struct with plan and usage
func (c *Client) GetAccount() (*Account, error) {
	account := &Account{}
	if _, err := c.doRequest("/account/", http.MethodGet, nil, nil, account); err != nil {
		return nil, err
	}
	return account, nil
}

// CreateDictionary creates a new dictionary using Dictionary struct properties
func (c *Client) CreateDictionary(d *Dictionary) (*HTTPResponse, error) {
	return c.doRequest("/entities/"+d.ID, http.MethodPut, DefaultHeaders(contentTypeJSON), d, &EmptyResponse{})
}

// GetDictionaries returns a list of all dictionaries
func (c *Client) GetDictionaries() (*HTTPResponse, error) { // FIXME: would be better to return a slice of Dictionary, but need to figured out how to keep the HTTPResponse reference
	return c.doRequest("/entities/", http.MethodGet, nil, nil, &EmptyResponse{})
}

// GetDictionary returns a Dictionary by id
func (c *Client) GetDictionary(ID string) (*Dictionary, error) {
	dict := &Dictionary{}
	if _, err := c.doRequest("/entities/"+ID, http.MethodGet, nil, nil, dict); err != nil {
		return nil, err
	}
	return dict, nil
}

// DeleteDictionary deletes a dictionary by id
func (c *Client) DeleteDictionary(ID string) (*HTTPResponse, error) {
	return c.doRequest("/entities/"+ID, http.MethodDelete, nil, nil, &EmptyResponse{})
}

// AddDictionaryEntries adds entries to a dictionary
func (c *Client) AddDictionaryEntries(ID string, e []DictionaryEntry) (*HTTPResponse, error) {
	return c.doRequest("/entities/"+ID+"/", http.MethodPost, DefaultHeaders(contentTypeJSON), &DictionaryEntryList{Entries: e}, &EmptyResponse{})
}

// AddDictionaryEntry adds an entry to a dictionary
func (c *Client) AddDictionaryEntry(ID string, e *DictionaryEntry) (*HTTPResponse, error) {
	return c.AddDictionaryEntries(ID, []DictionaryEntry{*e})
}

// GetDictionaryEntries returns a list of all entries for a dictionary
func (c *Client) GetDictionaryEntries(ID string, limit, offset int) (*DictionaryEntryList, error) { // FIXME: would be better to return a slice of Dictionary, but need to figured out how to keep the HTTPResponse reference
	params := Params{"limit": {string(limit)}, "offset": {string(offset)}}
	el := &DictionaryEntryList{}
	if _, err := c.doRequest("/entities/"+ID+"/_all", http.MethodGet, nil, params, el); err != nil {
		return nil, err
	}
	return el, nil
}

// GetDictionaryEntry returns a Dictionary Entry by id
func (c *Client) GetDictionaryEntry(dictID, entryID string) (*DictionaryEntry, error) {
	e := &DictionaryEntry{}
	if _, err := c.doRequest("/entities/"+dictID+"/"+entryID, http.MethodGet, nil, nil, e); err != nil {
		return nil, err
	}
	return e, nil
}

// DeleteDictionaryEntry deletes a Dictionary Entry by id
func (c *Client) DeleteDictionaryEntry(dictID, entryID string) (*HTTPResponse, error) {
	return c.doRequest("/entities/"+dictID+"/"+entryID, http.MethodDelete, nil, nil, &EmptyResponse{})
}

// CreateClassifierFromJSON creates a new classifier from a JSON string
func (c *Client) CreateClassifierFromJSON(ID, jsonStr string) (*HTTPResponse, error) {
	return c.doRequest("/categories/"+ID, http.MethodPut, DefaultHeaders(contentTypeJSON), &rawRequest{Body: jsonStr}, &EmptyResponse{})
}

// CreateClassifierFromCSV creates a new classifier from a CSV string
func (c *Client) CreateClassifierFromCSV(ID, csvStr string) (*HTTPResponse, error) {
	return c.doRequest("/categories/"+ID, http.MethodPut, DefaultHeaders(contentTypeCSV), &rawRequest{Body: csvStr}, &EmptyResponse{})
}

// DeleteClassifier deletes a Classifier by id
func (c *Client) DeleteClassifier(ID string) (*HTTPResponse, error) {
	return c.doRequest("/categories/"+ID, http.MethodDelete, nil, nil, &EmptyResponse{})
}

// GetClassifierCategories returns a list of all categories for a Classifier
func (c *Client) GetClassifierCategories(ID string, limit, offset int) (*CategoryList, error) {
	params := Params{"limit": {string(limit)}, "offset": {string(offset)}}
	cl := &CategoryList{}
	if _, err := c.doRequest("/categories/"+ID+"/_all", http.MethodGet, nil, params, cl); err != nil {
		return nil, err
	}
	return cl, nil
}

// GetClassifierCategory returns a Classifier Category by id
func (c *Client) GetClassifierCategory(clID, catID string) (*Category, error) {
	cat := &Category{}
	if _, err := c.doRequest("/categories/"+clID+"/"+catID, http.MethodGet, nil, nil, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

// DeleteClassifierCategory deletes a Classifier Category by id
func (c *Client) DeleteClassifierCategory(clID, catID string) (*HTTPResponse, error) {
	return c.doRequest("/categories/"+clID+"/"+catID, http.MethodDelete, nil, nil, &EmptyResponse{})
}
