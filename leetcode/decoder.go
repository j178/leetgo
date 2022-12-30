package leetcode

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"

	"github.com/dghubble/sling"
	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/utils"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/tidwall/gjson"
)

type smartDecoder struct {
	LogResponseData bool
	path            string
}

func (d smartDecoder) Decode(resp *http.Response, v interface{}) error {
	data, _ := io.ReadAll(resp.Body)
	dataStr := "<omitted>"
	if d.LogResponseData {
		dataStr = utils.BytesToString(data)
	}
	hclog.L().Trace("response", "url", resp.Request.URL.String(), "data", dataStr)

	ty := reflect.TypeOf(v)
	ele := reflect.ValueOf(v).Elem()
	switch ty.Elem() {
	case reflect.TypeOf(gjson.Result{}):
		if d.path == "" {
			ele.Set(reflect.ValueOf(gjson.ParseBytes(data)))
		} else {
			ele.Set(reflect.ValueOf(gjson.GetBytes(data, d.path)))
		}
	case reflect.TypeOf([]byte{}):
		ele.SetBytes(data)
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
