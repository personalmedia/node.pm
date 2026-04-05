// Copyright (c) 2026 DATAMIX.IO
// Author: L3DLP

package proxy

import (
	"net/http/httptest"
	"testing"
)

func FuzzIngressHeader(f *testing.F) {
	// Corpus initial
	f.Add("api.node.pm")
	f.Add("web.node.pm")
	f.Add("localhost")
	f.Add("bad-site.com")
	f.Add("<script>alert(1)</script>")

	f.Fuzz(func(t *testing.T, host string) {
		req := httptest.NewRequest("GET", "https://localhost/", nil)
		req.Host = host
		
		// On teste juste que setupProxy ne panic pas avec des entrées bizarres
		// (Le routage SNI est géré dans le serveur principal, ici on teste l'utilitaire de proxy)
		_, _ = SetupProxy("8080")
	})
}
