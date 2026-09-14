package efsapi

import "net/http"

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36"

// setBrowserHeaders decorates req with headers a real browser would send,
// so requests don't stand out from normal site traffic.
func setBrowserHeaders(req *http.Request, referer string) {
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-FR,fr;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Referer", referer)
	req.Header.Set("sec-ch-ua", `"Chromium";v="129", "Not=A?Brand";v="8", "Google Chrome";v="129"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
}
