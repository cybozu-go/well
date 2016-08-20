package cmd

import "time"

// AccessLog is to decode access log records from HTTPServer.
//
// The struct is tagged for JSON formatted records.
type AccessLog struct {
	Topic    string    `json:"topic"`
	LoggedAt time.Time `json:"logged_at"`
	Severity string    `json:"severity"`
	Utsname  string    `json:"utsname"`
	Message  string    `json:"message"`

	Type           string  `json:"type"`             // "access"
	Elapsed        float64 `json:"response_time"`    // floating point number of seconds.
	Protocol       string  `json:"protocol"`         // "HTTP/1.1" or alike
	StatusCode     int     `json:"http_status_code"` // 200, 404, ...
	Method         string  `json:"http_method"`
	RequestURI     string  `json:"url"`
	Host           string  `json:"http_host"`
	RequestLength  int64   `json:"request_size"`
	ResponseLength int64   `json:"response_size"`
	RemoteAddr     string  `json:"remote_ipaddr"`
	UserAgent      string  `json:"http_user_agent"`
	RequestID      string  `json:"id"`
}
