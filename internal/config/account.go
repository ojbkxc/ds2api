package config

import "strings"

func (a Account) Identifier() string {
	if strings.TrimSpace(a.Email) != "" {
		return strings.TrimSpace(a.Email)
	}
	if mobile := NormalizeMobileForStorage(a.Mobile); mobile != "" {
		return mobile
	}
	return ""
}

func (a Account) SupportsModel(model string) bool {
	if len(a.AllowedModels) == 0 {
		return true
	}
	lower := strings.ToLower(strings.TrimSpace(model))
	for _, m := range a.AllowedModels {
		if strings.ToLower(strings.TrimSpace(m)) == lower {
			return true
		}
	}
	return false
}
