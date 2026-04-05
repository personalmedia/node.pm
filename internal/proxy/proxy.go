// Copyright (c) 2026 DATAMIX.IO
// Author: L3DLP

package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// SetupProxy configure un proxy inverse vers la cible locale (ex: localhost:8080)
func SetupProxy(localTarget string) (http.Handler, error) {
	// Par défaut on cible localhost
	targetURL := fmt.Sprintf("http://127.0.0.1:%s", localTarget)
	u, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(u)

	// Modification du Header Host pour que l'application locale reçoive localhost
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = u.Host
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		req.Header.Set("X-Real-IP", req.RemoteAddr)
	}

	return proxy, nil
}
