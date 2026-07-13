package localtool

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileAuditResult struct {
	Allowed  bool              `json:"allowed"`
	FilePath string            `json:"file_path"`
	FileName string            `json:"file_name"`
	FileSize int64             `json:"file_size"`
	Hash     string            `json:"hash,omitempty"`
	Error    *ToolError        `json:"error,omitempty"`
	Details  map[string]string `json:"details,omitempty"`
}

type FileAuditor interface {
	AuditFile(filePath string, maxSizeBytes int64) *FileAuditResult
	AuditContent(content []byte, filename string, maxSizeBytes int64) *FileAuditResult
}

type DefaultFileAuditor struct {
	AllowedExtensions []string
	BlockedExtensions []string
	AllowedPaths      []string
	BlockedPaths      []string
	MaxFileSizeBytes  int64
}

func NewDefaultFileAuditor() *DefaultFileAuditor {
	return &DefaultFileAuditor{
		AllowedExtensions: []string{".txt", ".md", ".json", ".yaml", ".yml", ".csv", ".log", ".go", ".py", ".js", ".ts", ".html", ".css"},
		BlockedExtensions: []string{".exe", ".dll", ".bat", ".cmd", ".ps1", ".sh", ".apk", ".zip", ".rar", ".tar", ".gz"},
		MaxFileSizeBytes:  10 * 1024 * 1024,
	}
}

func (a *DefaultFileAuditor) AuditFile(filePath string, maxSizeBytes int64) *FileAuditResult {
	if maxSizeBytes <= 0 {
		maxSizeBytes = a.MaxFileSizeBytes
	}

	result := &FileAuditResult{
		FilePath: filePath,
		FileName: filepath.Base(filePath),
		Details:  make(map[string]string),
	}

	info, err := os.Stat(filePath)
	if err != nil {
		result.Allowed = false
		result.Error = &ToolError{Code: "file_not_found", Message: err.Error(), Retryable: false}
		return result
	}

	if info.IsDir() {
		result.Allowed = false
		result.Error = &ToolError{Code: "is_directory", Message: "path is a directory", Retryable: false}
		return result
	}

	result.FileSize = info.Size()
	if result.FileSize > maxSizeBytes {
		result.Allowed = false
		result.Error = &ToolError{Code: "file_too_large", Message: "file size exceeds limit", Retryable: false}
		result.Details["max_size"] = fmtIntBytes(maxSizeBytes)
		result.Details["actual_size"] = fmtIntBytes(result.FileSize)
		return result
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if len(a.BlockedExtensions) > 0 {
		for _, blocked := range a.BlockedExtensions {
			if strings.EqualFold(ext, blocked) {
				result.Allowed = false
				result.Error = &ToolError{Code: "blocked_extension", Message: "file extension is blocked", Retryable: false}
				result.Details["extension"] = ext
				return result
			}
		}
	}

	if len(a.AllowedExtensions) > 0 {
		allowed := false
		for _, allowedExt := range a.AllowedExtensions {
			if strings.EqualFold(ext, allowedExt) {
				allowed = true
				break
			}
		}
		if !allowed {
			result.Allowed = false
			result.Error = &ToolError{Code: "not_allowed_extension", Message: "file extension is not allowed", Retryable: false}
			result.Details["extension"] = ext
			return result
		}
	}

	if len(a.BlockedPaths) > 0 {
		for _, blocked := range a.BlockedPaths {
			if strings.Contains(strings.ToLower(filePath), strings.ToLower(blocked)) {
				result.Allowed = false
				result.Error = &ToolError{Code: "blocked_path", Message: "file path is blocked", Retryable: false}
				result.Details["blocked_pattern"] = blocked
				return result
			}
		}
	}

	if len(a.AllowedPaths) > 0 {
		allowed := false
		for _, allowedPath := range a.AllowedPaths {
			if strings.Contains(strings.ToLower(filePath), strings.ToLower(allowedPath)) {
				allowed = true
				break
			}
		}
		if !allowed {
			result.Allowed = false
			result.Error = &ToolError{Code: "not_allowed_path", Message: "file path is not allowed", Retryable: false}
			return result
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		result.Allowed = false
		result.Error = &ToolError{Code: "file_open_failed", Message: err.Error(), Retryable: false}
		return result
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		result.Allowed = false
		result.Error = &ToolError{Code: "hash_failed", Message: err.Error(), Retryable: false}
		return result
	}
	result.Hash = hex.EncodeToString(hash.Sum(nil))

	result.Allowed = true
	result.Details["audit_time"] = time.Now().Format(time.RFC3339)
	return result
}

func (a *DefaultFileAuditor) AuditContent(content []byte, filename string, maxSizeBytes int64) *FileAuditResult {
	if maxSizeBytes <= 0 {
		maxSizeBytes = a.MaxFileSizeBytes
	}

	result := &FileAuditResult{
		FilePath: filename,
		FileName: filepath.Base(filename),
		FileSize: int64(len(content)),
		Details:  make(map[string]string),
	}

	if result.FileSize > maxSizeBytes {
		result.Allowed = false
		result.Error = &ToolError{Code: "content_too_large", Message: "content size exceeds limit", Retryable: false}
		result.Details["max_size"] = fmtIntBytes(maxSizeBytes)
		result.Details["actual_size"] = fmtIntBytes(result.FileSize)
		return result
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if len(a.BlockedExtensions) > 0 {
		for _, blocked := range a.BlockedExtensions {
			if strings.EqualFold(ext, blocked) {
				result.Allowed = false
				result.Error = &ToolError{Code: "blocked_extension", Message: "file extension is blocked", Retryable: false}
				result.Details["extension"] = ext
				return result
			}
		}
	}

	if len(a.AllowedExtensions) > 0 {
		allowed := false
		for _, allowedExt := range a.AllowedExtensions {
			if strings.EqualFold(ext, allowedExt) {
				allowed = true
				break
			}
		}
		if !allowed {
			result.Allowed = false
			result.Error = &ToolError{Code: "not_allowed_extension", Message: "file extension is not allowed", Retryable: false}
			result.Details["extension"] = ext
			return result
		}
	}

	hash := sha256.Sum256(content)
	result.Hash = hex.EncodeToString(hash[:])

	result.Allowed = true
	result.Details["audit_time"] = time.Now().Format(time.RFC3339)
	return result
}

func fmtIntBytes(n int64) string {
	if n < 1024 {
		return fmtInt(int(n)) + " B"
	}
	if n < 1024*1024 {
		return fmtInt(int(n/1024)) + " KB"
	}
	return fmtInt(int(n/(1024*1024))) + " MB"
}
