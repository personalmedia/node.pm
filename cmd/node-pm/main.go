// Copyright (c) 2026 DATAMIX.IO
// Author: L3DLP

package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/getlantern/systray"
	"github.com/personalmedia/node.pm/internal/api"
	"github.com/personalmedia/node.pm/internal/proxy"
	"github.com/personalmedia/node.pm/internal/ui"
)

var (
	token      string
	tokenMu    sync.Mutex
	
	// État du Proxy Multi-Domaine Hardened
	certs       = make(map[string]*tls.Certificate)
	targets     = make(map[string]string)
	proxyMu     sync.RWMutex
	proxyServer *http.Server
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("Node.pm")
	systray.SetTooltip("Agent Node.pm - Sécurisé")

	mToken := systray.AddMenuItem("Saisir le Token", "Configurer l'agent")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quitter", "Fermer")

	go func() {
		for {
			select {
			case <-mToken.ClickedCh:
				newToken, ok, err := ui.EnterToken()
				if err == nil && ok {
					tokenMu.Lock()
					token = newToken
					tokenMu.Unlock()
					go doPing()
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			doPing()
		}
	}()
}

func onExit() {}

func doPing() {
	tokenMu.Lock()
	t := token
	tokenMu.Unlock()
	if t == "" { return }

	nodes, err := api.FetchNodes(t)
	if err != nil {
		log.Printf("Erreur API : %v", err)
		systray.SetTooltip("Erreur : " + err.Error())
		return
	}

	updateProxy(nodes)
}

func updateProxy(nodes []api.NodeInfo) {
	proxyMu.Lock()
	defer proxyMu.Unlock()

	newCerts := make(map[string]*tls.Certificate)
	newTargets := make(map[string]string)
	countOK := 0

	for _, node := range nodes {
		if node.Status == "OK" {
			// On force le domaine node.pm pour la sécurité et la marque
			fullDomain := fmt.Sprintf("%s.node.pm", node.Subdomain)
			
			if len(node.Certificate) < 64 || len(node.PrivateKey) < 64 {
				log.Printf("Certificat incomplet pour %s", fullDomain)
				continue
			}

			cert, err := tls.X509KeyPair([]byte(node.Certificate), []byte(node.PrivateKey))
			if err != nil {
				log.Printf("Erreur X509 pour %s : %v", fullDomain, err)
				continue
			}
			newCerts[fullDomain] = &cert
			newTargets[fullDomain] = node.LocalTarget
			countOK++
		}
	}

	certs = newCerts
	targets = newTargets

	if countOK > 0 {
		systray.SetTooltip(fmt.Sprintf("%d Nœuds Sécurisés", countOK))
		if proxyServer == nil {
			go startHttpsServer()
		}
	} else {
		systray.SetTooltip("En cours... (propagation DNS)")
	}
}

func startHttpsServer() {
	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			proxyMu.RLock()
			defer proxyMu.RUnlock()
			
			cert, ok := certs[info.ServerName]
			if !ok {
				return nil, fmt.Errorf("accès refusé pour SNI: %s", info.ServerName)
			}
			return cert, nil
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxyMu.RLock()
		targetPort, ok := targets[r.Host]
		proxyMu.RUnlock()

		if !ok {
			http.Error(w, "Node Not Found", 404)
			return
		}

		handler, err := proxy.SetupProxy(targetPort)
		if err != nil {
			http.Error(w, "Internal Proxy Error", 500)
			return
		}
		
		// Headers Hardened
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		
		handler.ServeHTTP(w, r)
	})

	proxyServer = &http.Server{
		Addr:         ":443",
		Handler:      mux,
		TLSConfig:    tlsCfg,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Println("Agent Node.pm HTTPS Proxy Actif sur :443")
	if err := proxyServer.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("Erreur fatale HTTPS : %v", err)
	}
}
