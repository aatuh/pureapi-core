package server

import "net/http"

// trackingResponseWriter wraps http.ResponseWriter to track header writes and bytes written.
type trackingResponseWriter struct {
	http.ResponseWriter
	wroteHeader  bool
	bytesWritten int64
}

// newTrackingResponseWriter creates a new tracking response writer.
func newTrackingResponseWriter(w http.ResponseWriter) *trackingResponseWriter {
	return &trackingResponseWriter{
		ResponseWriter: w,
	}
}

// WriteHeader records that headers have been written and calls the underlying WriteHeader.
func (w *trackingResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		// Headers already written, avoid double WriteHeader
		return
	}
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(code)
}

// Write records bytes written and calls the underlying Write.
func (w *trackingResponseWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(data)
	w.bytesWritten += int64(n)
	return n, err
}

// WroteHeader returns true if headers have been written.
func (w *trackingResponseWriter) WroteHeader() bool {
	return w.wroteHeader
}

// BytesWritten returns the number of bytes written to the response.
func (w *trackingResponseWriter) BytesWritten() int64 {
	return w.bytesWritten
}

// CanWriteHeader returns true if headers can still be written.
func (w *trackingResponseWriter) CanWriteHeader() bool {
	return !w.wroteHeader
}
