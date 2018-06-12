package gohttpmw

import "net/http"

type augmentedResponseWriter struct {
	http.ResponseWriter
	length     int
	httpStatus int
}

// WriteHeader will not only write b to w
// but also save the http status in the struct
func (w *augmentedResponseWriter) WriteHeader(httpStatus int) {
	w.ResponseWriter.WriteHeader(httpStatus)
	w.httpStatus = httpStatus
}

// Write will not only write b to w but also save the byte length in the struct
func (w *augmentedResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.length = n

	return n, err
}

func newAugmentedResponseWriter(
	w http.ResponseWriter,
) *augmentedResponseWriter {
	return &augmentedResponseWriter{
		ResponseWriter: w,
		httpStatus:     http.StatusOK,
	}
}
