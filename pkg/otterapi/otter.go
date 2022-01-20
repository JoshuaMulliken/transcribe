package otterapi

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

// getCSRFToken Gets the CSRF token from the login_csrf endpoint.
func getCSRFToken() (string, error) {
	// Create a new http request
	req, err := http.NewRequest("GET", "https://otter.ai/forward/api/v1/login_csrf", nil)
	if err != nil {
		panic(err)
	}

	// Get the response
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	// Get the cookies
	cookies := resp.Cookies()

	// Get the CSRF token
	for _, cookie := range cookies {
		if cookie.Name == "csrftoken" {
			return cookie.Value, nil
		}
	}

	// Get the response body
	body, err := getResponseBody(resp)
	if err != nil {
		return "", err
	}

	return "", errors.New(fmt.Sprintf("Unable to find CSRF token %s", body))
}

func getResponseBody(resp *http.Response) (string, error) {
	// Get the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// Login logs into the Otter API using a username and password.
func Login(username *string, password *string) (string, error) {
	// Url encode the username
	encodedUsername := url.QueryEscape(*username)

	// Create the uri for the request
	uri := fmt.Sprintf("https://otter.ai/forward/api/v1/login?username=%s", encodedUsername)

	// Create a new http request
	req, err := http.NewRequest("POST", uri, nil)
	if err != nil {
		panic(err)
	}

	// Get CSRF token
	csrfToken, err := getCSRFToken()
	if err != nil {
		panic(err)
	}

	req.AddCookie(&http.Cookie{
		Name:  "csrftoken",
		Value: csrfToken,
	})

	req.Header.Add("X-Csrftoken", csrfToken)

	// Add basic auth
	req.SetBasicAuth(*username, *password)

	// Get the response
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	// Get the session id from the cookies
	cookies := resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "sessionid" {
			return cookie.Value, nil
		}
	}

	return "", errors.New(fmt.Sprintf("Unable to find session id %s", cookies))
}

// getSpeechUploadParams Gets the parameters for the otter api.
func getSpeechUploadParams(sessionId string) (speechUploadParams, error) {
	// Create a new http request
	req, err := http.NewRequest("GET", "https://otter.ai/forward/api/v1/speech_upload_params", nil)
	if err != nil {
		panic(err)
	}

	// Add the session id
	req.AddCookie(&http.Cookie{
		Name:  "sessionid",
		Value: sessionId,
	})

	// Get the response
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	// Get the response body
	body, err := getResponseBody(resp)
	if err != nil {
		return speechUploadParams{}, err
	}

	// Parse the json response body
	params, err := parseSpeechUploadParams(body)
	if err != nil {
		return speechUploadParams{}, err
	}

	return params, nil
}

// speechUploadParams represents the parameters for the otter api as provided by amazon s3.
type speechUploadParams struct {
	Status string `json:"status"`
	Data   struct {
		AMZAlgorithm        string  `json:"x-amz-algorithm"`
		AMZSignature        string  `json:"x-amz-signature"`
		FormAction          string  `json:"form_action"`
		Key                 string  `json:"key"`
		AMZDate             string  `json:"x-amz-date"`
		Policy              string  `json:"policy"`
		AMZCredential       string  `json:"x-amz-credential"`
		SuccessActionStatus float64 `json:"success_action_status"`
		ACL                 string  `json:"acl"`
	}
}

// parseSpeechUploadParams parses the json response body into a speechUploadParams struct.
// This is a helper function for getSpeechUploadParams.
func parseSpeechUploadParams(body string) (speechUploadParams, error) {
	// Parse the json response body into a speechUploadParams struct
	var params speechUploadParams
	err := json.Unmarshal([]byte(body), &params)
	if err != nil {
		return params, err
	}

	return params, nil
}

// UploadSpeech uploads the audio file to amazon s3.
func UploadSpeech(sessionId string, audioFile *os.File) (string, error) {
	// Get the speech upload params
	params, err := getSpeechUploadParams(sessionId)
	if err != nil {
		return "", err
	}

	// Upload the file to amazon s3
	response, err := amazonBucketUpload(audioFile, params)
	if err != nil {
		return "", err
	}

	// Notify Otter that the file has been uploaded
	responseBody, err := notifyOtter(sessionId, response)
	if err != nil {
		return "", err
	}

	// Create the transcript URL
	transcriptUrl := fmt.Sprintf("https://otter.ai/u/%s", responseBody.OtID)

	return transcriptUrl, nil
}

// notifyOtter notifies Otter that the file has been uploaded.
func notifyOtter(sessionId string, response speechUploadResponse) (notifyOtterResponse, error) {
	// Query encode the Bucket and Key
	encodedBucket := url.QueryEscape(response.Bucket)
	encodedKey := url.QueryEscape(response.Key)

	// Create the uri for the request
	uri := fmt.Sprintf("https://otter.ai/forward/api/v1/finish_speech_upload?bucket=%s&key=%s&language=en&country=US&userid=821301", encodedBucket, encodedKey)

	// Create a new http request
	req, err := http.NewRequest("POST", uri, bytes.NewBufferString("{\n}"))
	if err != nil {
		panic(err)
	}

	// Add the session id
	req.AddCookie(&http.Cookie{
		Name:  "sessionid",
		Value: sessionId,
	})

	// Get a CSRF token
	csrfToken, err := getCSRFToken()
	if err != nil {
		panic(err)
	}

	// Add the CSRF token
	req.AddCookie(&http.Cookie{
		Name:  "csrftoken",
		Value: csrfToken,
	})
	req.Header.Add("X-Csrftoken", csrfToken)

	// Get the response
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	// Get the response body
	body, err := getResponseBody(resp)
	if err != nil {
		panic(err)
	}

	// Parse the json response body
	var responseBody notifyOtterResponse
	if err = json.Unmarshal([]byte(body), &responseBody); err != nil {
		panic(err)
	}

	return responseBody, nil
}

type notifyOtterResponse struct {
	Status   string  `json:"status"`
	SpeechID string  `json:"speech_id"`
	UploadID float64 `json:"upload_id"`
	OtID     string  `json:"otid"`
}

// amazonBucketUpload Uploads the file to the amazon bucket
func amazonBucketUpload(audioFile *os.File, params speechUploadParams) (speechUploadResponse, error) {
	// Create the multipart form data
	boundary, data, err := createSpeechFormData(audioFile, params)

	// Create a new http request
	req, err := http.NewRequest("POST", params.Data.FormAction, data)
	if err != nil {
		panic(err)
	}

	// Set the content type
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)

	// Get the response
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	// Get the response body
	body, err := getResponseBody(resp)
	if err != nil {
		return speechUploadResponse{}, err
	}

	// Parse the XML response body
	spchResponse, err := parseSpeechUploadResponse(body)
	if err != nil {
		return speechUploadResponse{}, err
	}

	return spchResponse, nil
}

type speechUploadResponse struct {
	XMLName  xml.Name `xml:"PostResponse"`
	Location string   `xml:"Location"`
	Bucket   string   `xml:"Bucket"`
	Key      string   `xml:"Key"`
	ETag     string   `xml:"ETag"`
}

// parseSpeechUploadResponse parses the XML response body into a speechUploadResponse struct.
// This is a helper function for amazonBucketUpload.
func parseSpeechUploadResponse(body string) (speechUploadResponse, error) {
	// Parse the XML response body into a speechUploadResponse struct
	var response speechUploadResponse
	err := xml.Unmarshal([]byte(body), &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

// createSpeechFormData creates the form data for uploading the audio file to amazon s3.
func createSpeechFormData(audioFile *os.File, params speechUploadParams) (string, io.Reader, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add the algorithm
	if err := w.WriteField("x-amz-algorithm", params.Data.AMZAlgorithm); err != nil {
		panic(err)
	}

	// Add the signature
	if err := w.WriteField("x-amz-signature", params.Data.AMZSignature); err != nil {
		panic(err)
	}

	// Add the key
	if err := w.WriteField("key", params.Data.Key); err != nil {
		panic(err)
	}

	// Add the date
	if err := w.WriteField("x-amz-date", params.Data.AMZDate); err != nil {
		panic(err)
	}

	// Add the policy
	if err := w.WriteField("policy", params.Data.Policy); err != nil {
		panic(err)
	}

	// Add the success action status
	if err := w.WriteField("success_action_status", fmt.Sprint(params.Data.SuccessActionStatus)); err != nil {
		panic(err)
	}

	// Add the credential
	if err := w.WriteField("x-amz-credential", params.Data.AMZCredential); err != nil {
		panic(err)
	}

	// Add the acl
	if err := w.WriteField("acl", params.Data.ACL); err != nil {
		panic(err)
	}

	// Add the file
	if fw, err := w.CreateFormFile("file", audioFile.Name()); err != nil {
		panic(err)
	} else {
		if _, err := io.Copy(fw, audioFile); err != nil {
			panic(err)
		}
	}

	// Get the boundary
	boundary := w.Boundary()

	// Close the writer
	err := w.Close()
	if err != nil {
		panic(err)
	}

	return boundary, &b, nil
}
