package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func main() {
	http.HandleFunc("/", handleRequest)
	http.ListenAndServe(":9000", nil)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	method := r.Method

	url, err := url.Parse(r.URL.String())
	if err != nil {
		fmt.Println("Error parsing URL: ", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	urlHostname := url.Hostname()
	url.Scheme = "https"
	url.Host = "api.openai.com"
	// 去掉 release 前缀（针对腾讯云）
	if strings.Contains(url.Path, "/release") {
		url.Path = strings.Replace(url.Path, "/release", "", 1)
	}

	requestHeaders := r.Header
	newRequestHeaders := make(http.Header)
	for key, values := range requestHeaders {
		for _, value := range values {
			newRequestHeaders.Add(key, value)
		}
	}
	newRequestHeaders.Set("Host", url.Host)
	newRequestHeaders.Set("Referer", url.Scheme+"://"+urlHostname)

	originalResponse, err := http.DefaultClient.Do(&http.Request{
		Method: method,
		URL:    url,
		Header: newRequestHeaders,
		Body:   r.Body,
	})
	if err != nil {
		fmt.Println("Error making request: ", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer originalResponse.Body.Close()

	originalResponseClone := originalResponse
	originalText, err := ioutil.ReadAll(originalResponseClone.Body)
	if err != nil {
		fmt.Println("Error reading response body: ", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responseHeaders := originalResponse.Header
	newResponseHeaders := make(http.Header)
	for key, values := range responseHeaders {
		for _, value := range values {
			newResponseHeaders.Add(key, value)
		}
	}
	status := originalResponse.StatusCode

	newResponseHeaders.Set("Cache-Control", "no-store")
	newResponseHeaders.Set("access-control-allow-origin", "*")
	newResponseHeaders.Set("access-control-allow-credentials", "true")
	newResponseHeaders.Del("content-security-policy")
	newResponseHeaders.Del("content-security-policy-report-only")
	newResponseHeaders.Del("clear-site-data")

	response := &http.Response{
		StatusCode: status,
		Header:     newResponseHeaders,
		Body:       ioutil.NopCloser(bytes.NewReader(originalText)),
	}

	for key, values := range response.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(response.StatusCode)
	io.Copy(w, response.Body)
}
