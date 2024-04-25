package migration

import (
	"bytes"
	"context"
	"net/http"
)

func DownloadFile(url string) (buf *bytes.Buffer, err error) {
	return DownloadFileWithHeaders(url, nil)
}

func DownloadFileWithHeaders(url string, headers http.Header) (buf *bytes.Buffer, err error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	for key, h := range headers {
		for _, hh := range h {
			req.Header.Add(key, hh)
		}
	}

	hc := http.Client{}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	buf = &bytes.Buffer{}
	_, err = buf.ReadFrom(resp.Body)

	return
}
