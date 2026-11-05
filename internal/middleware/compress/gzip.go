package compress

import (
	"compress/gzip"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type compressWriter struct {
	gin.ResponseWriter
	zw *gzip.Writer
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func Gzip() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.EqualFold(strings.TrimSpace(c.GetHeader("Content-Encoding")), "gzip") {
			reader, err := newCompressReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			defer reader.Close()

			c.Request.Body = reader
			c.Request.ContentLength = -1
			c.Request.Header.Del("Content-Encoding")
			c.Request.Header.Del("Content-Length")
		}

		if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			writer := &compressWriter{ResponseWriter: c.Writer}
			c.Writer = writer
			defer writer.Close()
		}
		c.Next()
	}
}

func (w *compressWriter) WriteHeaderNow() {
	if !w.Written() && w.canCompress() {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length")
		w.zw = gzip.NewWriter(w.ResponseWriter)
	}
	w.ResponseWriter.WriteHeaderNow()
}

func (w *compressWriter) Write(data []byte) (int, error) {
	w.WriteHeaderNow()
	if w.zw != nil {
		return w.zw.Write(data)
	}
	return w.ResponseWriter.Write(data)
}

func (w *compressWriter) WriteString(data string) (int, error) {
	return w.Write([]byte(data))
}

func (w *compressWriter) Close() error {
	if w.zw != nil {
		return w.zw.Close()
	}
	return nil
}

func (r *compressReader) Read(p []byte) (int, error) {
	return r.zr.Read(p)
}

func (r *compressReader) Close() error {
	return errors.Join(r.zr.Close(), r.r.Close())
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &compressReader{r: r, zr: zr}, nil
}

func (w *compressWriter) canCompress() bool {
	mediaType, _, _ := mime.ParseMediaType(w.Header().Get("Content-Type"))
	return mediaType == "application/json" || mediaType == "text/html"
}
