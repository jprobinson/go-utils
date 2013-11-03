package web

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"github.com/jprobinson/go-utils/utils"
)

// AccessLogHandler return a http.Handler that wraps h and logs requests to out in
// Apache Combined Log Format...with an NYT twist (SRCIP instead of RemoteAddr if it exists).
//
// See http://httpd.apache.org/docs/2.2/logs.html#combined for a description of this format.
//
// AccessLogHandler always sets the ident field of the log to -
func AccessLogHandler(access_log_name string, h http.Handler) http.Handler {
	log_handler := accessLogHandler{utils.GetLogFileHandle(access_log_name), h, access_log_name}
	go utils.ListenForLogSignal(&log_handler)
	return &log_handler
}

// SetupLogging implements the utils.LogSetup interface so we can use the utils package to
// detect SIGHUP signals to alert us when logrotate is complete.
func (h *accessLogHandler) SetupLogging() {
	h.writer = utils.GetLogFileHandle(h.logFile)
	fmt.Fprintf(h.writer, "setup new access logger")
}

// accessLogHandler is the http.Handler implementation for creating an access log
type accessLogHandler struct {
	writer  io.Writer
	handler http.Handler
	logFile string
}

func (h *accessLogHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	t := time.Now()
	logger := responseLogger{w: w}
	h.handler.ServeHTTP(&logger, req)
	writeAccessLog(h.writer, req, t, logger.status, logger.size)
}

// responseLogger is wrapper of http.ResponseWriter that keeps track of its HTTP status
// code and body size
type responseLogger struct {
	w      http.ResponseWriter
	status int
	size   int
}

func (l *responseLogger) Header() http.Header {
	return l.w.Header()
}

func (l *responseLogger) Write(b []byte) (int, error) {
	if l.status == 0 {
		// The status will be StatusOK if WriteHeader has not been called yet
		l.status = http.StatusOK
	}
	size, err := l.w.Write(b)
	l.size += size
	return size, err
}

func (l *responseLogger) WriteHeader(s int) {
	l.w.WriteHeader(s)
	l.status = s
}

// buildAccessLogLine builds a log entry for req based on Apache Common Log Format.
// ts is the timestamp with which the entry should be logged.
// status and size are used to provide the response HTTP status and size.
func buildAccessLogLine(req *http.Request, ts time.Time, status int, size int) string {
	return fmt.Sprintf("%s - [%s] \"%s %s %s\" %d %d \"%s\" \"%s\"\n",
		getSourceIP(req),
		ts.Format("02/Jan/2006:15:04:05 -0700"),
		req.Method,
		strings.Replace(req.RequestURI, "%", "%%", -1),
		req.Proto,
		status,
		size,
		req.Referer(),
		strings.Replace(req.UserAgent(), "%", "%%", -1))
}

func getSourceIP(req *http.Request) (ip string) {
	if ip = req.Header.Get("SRCIP"); ip == "" {
		ip = strings.Split(req.RemoteAddr, ":")[0]
	}
	return ip
}

// writeAccessLog writes a log entry for req to w in Apache Combined Log Format...with an NYT twist (SRCIP instead of RemoteAddr if it exists).
// ts is the timestamp with which the entry should be logged.
// status and size are used to provide the response HTTP status and size.
func writeAccessLog(w io.Writer, req *http.Request, ts time.Time, status, size int) {
	line := buildAccessLogLine(req, ts, status, size)
	fmt.Fprintf(w, line)
}
