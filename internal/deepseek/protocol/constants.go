package protocol

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

const (
	DeepSeekHost                 = "chat.deepseek.com"
	DeepSeekLoginURL             = "https://chat.deepseek.com/api/v0/users/login"
	DeepSeekCreateSessionURL     = "https://chat.deepseek.com/api/v0/chat_session/create"
	DeepSeekCreatePowURL         = "https://chat.deepseek.com/api/v0/chat/create_pow_challenge"
	DeepSeekCompletionURL        = "https://chat.deepseek.com/api/v0/chat/completion"
	DeepSeekContinueURL          = "https://chat.deepseek.com/api/v0/chat/continue"
	DeepSeekUploadFileURL        = "https://chat.deepseek.com/api/v0/file/upload_file"
	DeepSeekFetchFilesURL        = "https://chat.deepseek.com/api/v0/file/fetch_files"
	DeepSeekFetchSessionURL      = "https://chat.deepseek.com/api/v0/chat_session/fetch_page"
	DeepSeekDeleteSessionURL     = "https://chat.deepseek.com/api/v0/chat_session/delete"
	DeepSeekDeleteAllSessionsURL = "https://chat.deepseek.com/api/v0/chat_session/delete_all"
	DeepSeekCompletionTargetPath = "/api/v0/chat/completion"
	DeepSeekUploadTargetPath     = "/api/v0/file/upload_file"
)

var defaultStaticBaseHeaders = map[string]string{
	"Host":           "chat.deepseek.com",
	"Accept":         "application/json",
	"Content-Type":   "application/json",
	"accept-charset": "UTF-8",
}

var defaultSkipContainsPatterns = []string{
	"quasi_status",
	"elapsed_secs",
	"token_usage",
	"pending_fragment",
	"conversation_mode",
	"fragments/-1/status",
	"fragments/-2/status",
	"fragments/-3/status",
}

var defaultSkipExactPaths = []string{
	"response/search_status",
}

var ClientVersion string
var BaseHeaders = map[string]string{}
var SkipContainsPatterns = cloneStringSlice(defaultSkipContainsPatterns)
var SkipExactPathSet = toStringSet(defaultSkipExactPaths)

type clientConstants struct {
	Name            string `json:"name"`
	Platform        string `json:"platform"`
	Version         string `json:"version"`
	AndroidAPILevel string `json:"android_api_level"`
	Locale          string `json:"locale"`
}

type sharedConstants struct {
	Client              clientConstants   `json:"client"`
	BaseHeaders         map[string]string `json:"base_headers"`
	SkipContainsPattern []string          `json:"skip_contains_patterns"`
	SkipExactPaths      []string          `json:"skip_exact_paths"`
}

//go:embed constants_shared.json
var sharedConstantsJSON []byte

func init() {
	cfg := sharedConstants{}
	if err := json.Unmarshal(sharedConstantsJSON, &cfg); err != nil {
		panic(fmt.Errorf("load DeepSeek shared constants: %w", err))
	}
	applySharedConstants(cfg)
}

func applySharedConstants(cfg sharedConstants) {
	client := normalizeClientConstants(cfg.Client)
	ClientVersion = client.Version
	BaseHeaders = buildBaseHeaders(client, cfg.BaseHeaders)
	SkipContainsPatterns = cloneStringSlice(defaultSkipContainsPatterns)
	if len(cfg.SkipContainsPattern) > 0 {
		SkipContainsPatterns = cloneStringSlice(cfg.SkipContainsPattern)
	}
	SkipExactPathSet = toStringSet(defaultSkipExactPaths)
	if len(cfg.SkipExactPaths) > 0 {
		SkipExactPathSet = toStringSet(cfg.SkipExactPaths)
	}
}

func normalizeClientConstants(in clientConstants) clientConstants {
	if in.Name == "" {
		in.Name = "DeepSeek"
	}
	if in.Platform == "" {
		in.Platform = "android"
	}
	if in.AndroidAPILevel == "" {
		in.AndroidAPILevel = "35"
	}
	if in.Locale == "" {
		in.Locale = "zh_CN"
	}
	return in
}

func buildBaseHeaders(client clientConstants, overrides map[string]string) map[string]string {
	out := cloneStringMap(defaultStaticBaseHeaders)
	for k, v := range overrides {
		if k == "" || v == "" {
			continue
		}
		out[k] = v
	}
	if client.Name != "" && client.Version != "" {
		userAgent := client.Name + "/" + client.Version
		if client.Platform == "android" && client.AndroidAPILevel != "" {
			userAgent += " Android/" + client.AndroidAPILevel
		}
		out["User-Agent"] = userAgent
	}
	if client.Platform != "" {
		out["x-client-platform"] = client.Platform
	}
	if client.Version != "" {
		out["x-client-version"] = client.Version
	}
	if client.Locale != "" {
		out["x-client-locale"] = client.Locale
	}
	return out
}

func cloneStringMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneStringSlice(in []string) []string {
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func toStringSet(in []string) map[string]struct{} {
	out := make(map[string]struct{}, len(in))
	for _, v := range in {
		if v == "" {
			continue
		}
		out[v] = struct{}{}
	}
	return out
}

const (
	KeepAliveTimeout  = 5
	StreamIdleTimeout = 300
	MaxKeepaliveCount = 40
)

// Client variants to avoid all accounts using identical client fingerprints.
// Selected deterministically per account based on a hash of the account ID.
var clientVariants = []clientConstants{
	{Name: "DeepSeek", Platform: "android", Version: "2.1.0", AndroidAPILevel: "33", Locale: "zh_CN"},
	{Name: "DeepSeek", Platform: "android", Version: "2.2.0", AndroidAPILevel: "34", Locale: "zh_CN"},
	{Name: "DeepSeek", Platform: "android", Version: "2.2.0", AndroidAPILevel: "35", Locale: "zh_CN"},
	{Name: "DeepSeek", Platform: "android", Version: "2.1.1", AndroidAPILevel: "33", Locale: "zh_CN"},
	{Name: "DeepSeek", Platform: "android", Version: "2.3.0", AndroidAPILevel: "35", Locale: "zh_CN"},
}

// SelectClientVariant returns a client profile for the given account identifier.
// Uses a hash of the ID to deterministically pick a variant, ensuring each
// account consistently uses the same client fingerprint.
func SelectClientVariant(accountID string) clientConstants {
	if accountID == "" {
		return clientVariants[0]
	}
	h := fnvHash(accountID)
	return clientVariants[h%uint32(len(clientVariants))]
}

// BuildAccountHeaders produces per-account request headers, overriding the
// global base headers with the account-specific client variant.
func BuildAccountHeaders(accountID string, baseHeaders map[string]string) map[string]string {
	variant := SelectClientVariant(accountID)
	return buildBaseHeaders(variant, baseHeaders)
}

func fnvHash(s string) uint32 {
	const (
		offset32 = 2166136261
		prime32  = 16777619
	)
	h := uint32(offset32)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= prime32
	}
	return h
}
