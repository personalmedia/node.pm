// Copyright (c) 2026 DATAMIX.IO
// Author: L3DLP

package ui

import (
	"github.com/gen2brain/dlgs"
)

// EnterToken ouvre une boîte de dialogue native pour saisir le token de sécurité.
func EnterToken() (string, bool, error) {
	passwd, ok, err := dlgs.Password("Saisie Securisée", "Veuillez entrer votre Token Node.pm :")
	if err != nil {
		return "", false, err
	}
	return passwd, ok, nil
}
