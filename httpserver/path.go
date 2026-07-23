package httpserver

import "strings"

// joinPathSegments 将多段 path 合并为相对路径（无首尾 `/`），忽略空段。
// 例如："/a/", "", "b/:id/" → "a/b/:id"
func joinPathSegments(parts ...string) string {
	var segs []string
	for _, part := range parts {
		for _, seg := range strings.Split(part, "/") {
			if seg == "" {
				continue
			}
			segs = append(segs, seg)
		}
	}
	return strings.Join(segs, "/")
}

// joinHTTPPaths 将多段 path 合并为绝对路径，无尾斜杠（根路径除外）。
// 空 path / 仅斜杠不会产生尾斜杠，避免 WithPathRegister 收到 "/api/" 这类路径。
func joinHTTPPaths(parts ...string) string {
	s := joinPathSegments(parts...)
	if s == "" {
		return "/"
	}
	return "/" + s
}

// ginPathToOpenAPIPath 将 Gin 风格 :id / *path 转为 OpenAPI 风格 {id} / {path}。
func ginPathToOpenAPIPath(path string) string {
	segs := strings.Split(path, "/")
	for i, seg := range segs {
		switch {
		case strings.HasPrefix(seg, ":"):
			segs[i] = "{" + strings.TrimPrefix(seg, ":") + "}"
		case strings.HasPrefix(seg, "*"):
			segs[i] = "{" + strings.TrimPrefix(seg, "*") + "}"
		}
	}
	return strings.Join(segs, "/")
}
