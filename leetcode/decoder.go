package leetcode

import (
	"compress/gzip"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/dghubble/sling"
	"github.com/goccy/go-json"
	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/utils"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/tidwall/gjson"
)

var (
	gjsonType      = reflect.TypeOf(gjson.Result{})
	bytesType      = reflect.TypeOf([]byte{})
	stringType     = reflect.TypeOf("")
	errHandlerType = reflect.TypeOf(&defaultErrorHandler{})
)

type smartDecoder struct {
	Debug       bool
	LogResponse bool
	LogLimit    int
	path        string
}

func headerString(h http.Header) string {
	w := &strings.Builder{}
	_ = h.WriteSubset(
		w, map[string]bool{
			"Content-Security-Policy":          true,
			"Set-Cookie":                       true,
			"X-Frame-Options":                  true,
			"Vary":                             true,
			"Strict-Transport-Security":        true,
			"Date":                             true,
			"Access-Control-Allow-Credentials": true,
			"Access-Control-Allow-Origin":      true,
		},
	)
	return w.String()
}

func (d smartDecoder) Decode(resp *http.Response, v interface{}) error {
	if strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") && resp.ContentLength != 0 {
		if _, ok := resp.Body.(*gzip.Reader); !ok {
			var err error
			resp.Body, err = gzip.NewReader(resp.Body)
			if err != nil {
				return err
			}
		}
	}

	data, _ := io.ReadAll(resp.Body)

	if d.Debug {
		dataStr := "<omitted>"
		if d.LogResponse {
			dataStr = utils.BytesToString(data)
			limit := d.LogLimit
			if len(data) < limit {
				limit = len(data)
			}
			dataStr = dataStr[:limit]
		}
		hclog.L().Trace(
			"response",
			"url", resp.Request.URL.String(),
			"code", resp.StatusCode,
			"headers", headerString(resp.Header),
			"data", dataStr,
		)
	}

	ty := reflect.TypeOf(v)
	ele := reflect.ValueOf(v).Elem()
	switch ty.Elem() {
	case gjsonType:
		if d.path == "" {
			ele.Set(reflect.ValueOf(gjson.ParseBytes(data)))
		} else {
			ele.Set(reflect.ValueOf(gjson.GetBytes(data, d.path)))
		}
	case bytesType:
		ele.SetBytes(data)
	case stringType:
		ele.SetString(utils.BytesToString(data))
	case errHandlerType:
		ele.Set(reflect.ValueOf(&defaultErrorHandler{utils.BytesToString(data)}))
	default:
		return json.Unmarshal(data, v)
	}
	return nil
}

// It's proxy reader, implement io.Reader
type reader struct {
	io.Reader
	tracker *progress.Tracker
}

func (r *reader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.tracker.Increment(int64(n))
	return
}

// Close the reader when it implements io.Closer
func (r *reader) Close() (err error) {
	r.tracker.MarkAsDone()
	if closer, ok := r.Reader.(io.Closer); ok {
		return closer.Close()
	}
	return
}

type progressDecoder struct {
	sling.ResponseDecoder
	tracker *progress.Tracker
}

func (d progressDecoder) Decode(resp *http.Response, v interface{}) error {
	total := resp.ContentLength
	d.tracker.UpdateTotal(total)
	resp.Body = &reader{resp.Body, d.tracker}
	return d.ResponseDecoder.Decode(resp, v)
}
