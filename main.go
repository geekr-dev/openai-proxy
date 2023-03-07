package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	http.HandleFunc("/", handleRequest)
	http.ListenAndServe(":9000", nil)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	method := r.Method

	reqUrl, err := url.Parse(r.URL.String())
	if err != nil {
		fmt.Println("Error parsing URL: ", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	urlHostname := reqUrl.Hostname()
	reqUrl.Scheme = "https"
	reqUrl.Host = "api.openai.com"

	// 去掉环境前缀（针对腾讯云，如果包含的话，目前我只用到了test和release）
	reqUrl.Path = strings.Replace(reqUrl.Path, "/release", "", 1)
	reqUrl.Path = strings.Replace(reqUrl.Path, "/test", "", 1)

	// 请求头处理
	reqHeaders := r.Header
	newReqHeaders := make(http.Header)
	for key, values := range reqHeaders {
		for _, value := range values {
			newReqHeaders.Add(key, value)
		}
	}
	newReqHeaders.Set("Host", reqUrl.Host)
	newReqHeaders.Set("Referer", reqUrl.Scheme+"://"+urlHostname)

	// 超时时间设置为30s
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	// 本地测试通过代理请求 OpenAI 接口
	if os.Getenv("ENV") == "local" {
		proxyURL, _ := url.Parse("http://127.0.0.1:10809")
		client.Transport = &http.Transport{
			Proxy:           http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	var originResp *http.Response
	// 支持最大重试次数为3次
	tries := 0
	for tries < 3 {
		// 将request的body复制一份，以便在重试时使用
		var reqBody io.ReadCloser
		if r.Body != nil {
			bodyCopy, _ := ioutil.ReadAll(r.Body)
			reqBody = ioutil.NopCloser(bytes.NewBuffer(bodyCopy))
		}
		originResp, err = client.Do(&http.Request{
			Method: method,
			URL:    reqUrl,
			Header: newReqHeaders,
			Body:   reqBody,
		})
		if err != nil {
			fmt.Println("Error making request: ", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		// 状态码为429/433时，自动重试，否则退出
		if originResp.StatusCode == http.StatusTooManyRequests || originResp.StatusCode == 433 {
			tries++
			time.Sleep(1 * time.Second) // 间隔1s后重试
		} else {
			break
		}
	}
	defer originResp.Body.Close()

	// 处理响应数据返回给客户端
	originRespClone := originResp
	originalText, err := ioutil.ReadAll(originRespClone.Body)
	if err != nil {
		fmt.Println("Error reading response body: ", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	// 响应头
	responseHeaders := originResp.Header
	newRespHeaders := make(http.Header)
	for key, values := range responseHeaders {
		for _, value := range values {
			newRespHeaders.Add(key, value)
		}
	}
	newRespHeaders.Set("Cache-Control", "no-store")
	newRespHeaders.Set("access-control-allow-origin", "*")
	newRespHeaders.Set("access-control-allow-credentials", "true")
	newRespHeaders.Del("content-security-policy")
	newRespHeaders.Del("content-security-policy-report-only")
	newRespHeaders.Del("clear-site-data")

	// 响应对象封装
	response := &http.Response{
		StatusCode: originResp.StatusCode,
		Header:     newRespHeaders,
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
