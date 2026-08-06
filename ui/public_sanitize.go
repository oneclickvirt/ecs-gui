package ui

import (
	"encoding/json"
	"regexp"
	"strings"
)

var (
	guiURLPattern    = regexp.MustCompile(`(?i)\b(?:https?|ftp|ssh|git)://[^\s)]+`)
	guiGitPattern    = regexp.MustCompile(`(?i)\bgit@[^\s:]+:[^\s]+`)
	guiRepoPattern   = regexp.MustCompile(`(?i)\b(?:github\.com|gitlab\.com|bitbucket\.org)/[^\s/]+/[^\s]+`)
	guiSecretPattern = regexp.MustCompile(`(?i)(authorization|bearer|token|api[_-]?key|secret|password|passwd|auth)\s*[:=]\s*(?:bearer\s+)?["']?[^\s,;"'&#]+`)
	guiPathPattern   = regexp.MustCompile(`(?:^|\s)(?:/Users/|/Volumes/|/home/|/root/|[A-Za-z]:\\)[^\s]+`)
	guiQuerySecret   = regexp.MustCompile(`(?i)([?&](?:token|key|secret|password|passwd|auth)[^=]*=)[^&#\s]+`)
)

func sanitizeGUIText(value string) string {
	lines := strings.Split(value, "\n")
	for index, line := range lines {
		// Credentials must be removed even from an ordinary successful target
		// line. The target URL itself remains visible unless this is a loader or
		// error diagnostic.
		line = redactGUICredentials(line)
		lower := strings.ToLower(line)
		if !guiURLPattern.MatchString(line) && !guiGitPattern.MatchString(line) &&
			!guiRepoPattern.MatchString(line) &&
			!guiPathPattern.MatchString(line) {
			lines[index] = line
			continue
		}
		if containsGUIKeyword(lower,
			"registry", "manifest", "fallback", "data source", "datasource", "source url", "source:",
			"request", "endpoint", "fetch", "download", "load", "error", "failed", "warning",
			"数据源", "来源", "回退", "清单", "请求", "接口", "下载", "加载", "错误", "失败", "警告",
		) {
			line = redactGUISecrets(line)
		}
		lines[index] = line
	}
	return strings.Join(lines, "\n")
}

func redactGUICredentials(value string) string {
	value = guiQuerySecret.ReplaceAllString(value, "$1[redacted]")
	return guiSecretPattern.ReplaceAllString(value, "$1=[redacted]")
}

func redactGUISecrets(value string) string {
	value = redactGUICredentials(value)
	value = guiURLPattern.ReplaceAllString(value, "[remote-url]")
	value = guiGitPattern.ReplaceAllString(value, "[remote-source]")
	value = guiRepoPattern.ReplaceAllString(value, "[remote-source]")
	value = guiSecretPattern.ReplaceAllString(value, "$1=[redacted]")
	value = guiPathPattern.ReplaceAllString(value, " [local-path]")
	return value
}

func containsGUIKeyword(value string, keywords ...string) bool {
	for _, keyword := range keywords {
		if strings.Contains(value, keyword) {
			return true
		}
	}
	return false
}

func sanitizeStructuredRunResult(report *StructuredRunResult) {
	if report == nil {
		return
	}
	report.Data = nil
	report.DataFiles = nil
	report.Text = sanitizeGUIText(report.Text)
	for index := range report.Sections {
		report.Sections[index].Reason = safeStructuredReason(report.Sections[index].Status, report.Sections[index].Reason)
	}
	for index := range report.Components {
		report.Components[index].Reason = safeStructuredReason(report.Components[index].Status, report.Components[index].Reason)
		report.Components[index].Payload = sanitizeGUIPayloadForComponent(report.Components[index].Payload, report.Components[index].Name)
	}
}

func sanitizeGUIPayload(payload json.RawMessage) json.RawMessage {
	return sanitizeGUIPayloadForComponent(payload, "")
}

func sanitizeGUIPayloadForComponent(payload json.RawMessage, componentName string) json.RawMessage {
	if len(payload) == 0 {
		return payload
	}
	var value any
	if err := json.Unmarshal(payload, &value); err != nil {
		return json.RawMessage(`null`)
	}
	scope := ""
	if strings.Contains(strings.ToLower(componentName), "registry") {
		scope = "provenance"
	}
	sanitizeGUIJSONValueScoped(value, scope)
	encoded, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage(`null`)
	}
	return encoded
}

func sanitizeGUIJSONValue(value any) {
	sanitizeGUIJSONValueScoped(value, "")
}

func sanitizeGUIJSONValueScoped(value any, endpointScope string) {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			normalized := normalizeGUIJSONKey(key)
			if isGUIProvenanceJSONKey(normalized, endpointScope) || isGUICredentialJSONKey(normalized) ||
				(endpointScope != "" && isGUIEndpointJSONKey(normalized)) {
				delete(typed, key)
				continue
			}
			childScope := endpointScope
			if normalized == "rdap" || normalized == "whois" || normalized == "geofeed" || normalized == "geofeeds" {
				childScope = normalized
			}
			if isGUIProvenanceContainer(normalized) {
				childScope = "provenance"
			}
			if text, ok := child.(string); ok {
				if isGUIErrorJSONKey(normalized) {
					typed[key] = redactGUISecrets(text)
				} else if guiURLPattern.MatchString(text) || guiGitPattern.MatchString(text) || guiRepoPattern.MatchString(text) || guiPathPattern.MatchString(text) || guiSecretPattern.MatchString(text) {
					typed[key] = redactGUICredentials(text)
				}
				continue
			}
			sanitizeGUIJSONValueScoped(child, childScope)
		}
	case []any:
		for _, child := range typed {
			sanitizeGUIJSONValueScoped(child, endpointScope)
		}
	}
}

func normalizeGUIJSONKey(key string) string {
	return strings.NewReplacer("_", "", "-", "", ".", "").Replace(strings.ToLower(strings.TrimSpace(key)))
}

func isGUIProvenanceJSONKey(normalized, scope string) bool {
	switch normalized {
	case "privateregistry", "datafiles", "datasource", "sourceurl", "fallbackurl", "manifesturl", "registryurl",
		"registrysource", "registryfallback", "geofeedurl", "geofeedurls", "whoisserver", "whoisurl", "whoisendpoint",
		"rdapserver", "rdapurl", "rdapendpoint":
		return true
	case "source", "fallback", "manifest", "registry":
		return scope == "provenance"
	case "url":
		return scope == "provenance"
	}
	return (strings.HasPrefix(normalized, "geofeed") || strings.HasPrefix(normalized, "whois") || strings.HasPrefix(normalized, "rdap")) &&
		containsGUIKeyword(normalized, "url", "uri", "endpoint", "server")
}

func isGUIProvenanceContainer(normalized string) bool {
	return normalized == "registry" || normalized == "registryreport" || normalized == "registryresolution" ||
		normalized == "serverregistry" || normalized == "providermetadata" || normalized == "manifest"
}

func isGUIEndpointJSONKey(normalized string) bool {
	return normalized == "url" || normalized == "uri" || normalized == "href" || normalized == "server" ||
		normalized == "endpoint" || normalized == "baseurl"
}

func isGUICredentialJSONKey(normalized string) bool {
	return normalized == "authorization" || normalized == "bearer" || normalized == "token" || normalized == "apikey" ||
		normalized == "secret" || normalized == "password" || normalized == "passwd" || normalized == "credential" ||
		strings.HasSuffix(normalized, "token") || strings.HasSuffix(normalized, "apikey") || strings.HasSuffix(normalized, "secret") ||
		strings.HasSuffix(normalized, "password") || strings.HasSuffix(normalized, "credential")
}

func isGUIErrorJSONKey(normalized string) bool {
	return normalized == "error" || normalized == "reason" || normalized == "message" || normalized == "detail" ||
		strings.HasSuffix(normalized, "error") || strings.HasSuffix(normalized, "reason")
}

func safeStructuredReason(status, reason string) string {
	status = strings.TrimSpace(status)
	reason = strings.TrimSpace(reason)
	if reason == "" || reason == status {
		return ""
	}
	if containsGUIKeyword(strings.ToLower(reason), "registry", "manifest", "fallback", "data source", "datasource", "source url", "load ", "fetch ", "download ", "清单", "数据源", "加载", "获取", "下载") {
		return safeStatusReason(status)
	}
	return sanitizeGUIText(reason)
}

func safeStatusReason(status string) string {
	switch status {
	case "timeout":
		return "timeout"
	case "canceled":
		return "canceled"
	case "unavailable":
		return "unavailable"
	case "partial":
		return "partial"
	case "error":
		return "error"
	default:
		return "unavailable"
	}
}
