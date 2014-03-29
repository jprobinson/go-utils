package web

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const (
	dateISOFormat = "2006-01-02"
)

var JsonContentType = "application/json; charset=UTF-8"

// JsonResponse is a convenience type for generating JSON responses
type JsonResponse map[string]interface{}

type JsonResponseWrapper struct {
	Response interface{}
}

func AddStatusHandler(router *mux.Router) {
	router.HandleFunc("/status.txt", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprint(rw, "alive")
		return
	})
	return
}

func (r JsonResponse) String() (s string) {
	b, err := json.Marshal(r)
	if err != nil {
		s = ""
		return
	}
	s = string(b)
	return
}

func (r JsonResponseWrapper) String() (s string) {
	b, err := json.Marshal(r.Response)
	if err != nil {
		log.Printf("error converting response to json!!! - %s", err.Error())
		s = ""
		return
	}
	s = string(b)
	return
}

// An http.Handler which handles a certain prefix of URLs.
type PrefixHandler interface {
	UrlPrefix() string
	Handle(subRouter *mux.Router)
}

func getHandlerPath(ph PrefixHandler, handlerCommand string) string {
	return fmt.Sprintf("%s%s", ph.UrlPrefix(), handlerCommand)
}

func ServiceUrlNotFoundResponse(w http.ResponseWriter, url string) {
	fmt.Fprint(w, fmt.Sprintf("Service URL (%s) not found", url), http.StatusNotFound)
}

// Generic error response. This should be used for all errors if possible.
func ErrorResponse(w http.ResponseWriter, err error, status int) {
	w.Header().Set("Content-Type", JsonContentType)
	w.WriteHeader(status)
	fmt.Fprint(w, JsonResponse{"error": err.Error(), "status": "error"})
}

// Generic success response. This should be used in the case there is no
// additional information to report other than "ok". For APIs that are
// returning info, we should still append "status": "ok" to the response.
func SuccessResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", JsonContentType)
	fmt.Fprint(w, JsonResponse{"status": "ok"})
}

// ParseISODate will parse a YYYY-MM-DD string into a time.Time value.
func ParseISODate(dateStr string) (date time.Time, err error) {
	date, err = time.Parse(dateISOFormat, dateStr)
	return
}

// ParseDateRange will look for and parse 'startDate' and 'endDate' ISO date strings in the
// given vars map.
func ParseDateRange(vars map[string]string) (startDate time.Time, endDate time.Time, err error) {
	startDate, err = ParseISODate(vars["start"])
	if err != nil {
		err = errors.New("please use a valid start date with a format of YYYY-MM-DD")
		return
	}

	endDate, err = ParseISODate(vars["end"])
	if err != nil {
		err = errors.New("please use a valid end date with a format of YYYY-MM-DD")
		return
	}

	return
}

// ParseDateRangeFullDay will look for and parse 'startDate' and 'endDate' ISO date strings in the
// given vars map. It will then set the startDate time to midnight and the endDate time to 23:59:59.
func ParseDateRangeFullDay(vars map[string]string) (startDate time.Time, endDate time.Time, err error) {
	startDate, endDate, err = ParseDateRange(vars)
	if err != nil {
		return
	}

	// set time to beginning of day
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0,
		0, 0, time.Local)
	// set the time to the end of day
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59,
		59, 1000, time.Local)
	return
}

// Serialize encodes a value using gob.
func Serialize(src interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(src)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Deserialize decodes a value using gob.
func Deserialize(src []byte, dst interface{}) error {
	buf := bytes.NewBuffer(src)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(dst)
	if err != nil {
		return err
	}
	return nil
}

func DebugWebInput(vars map[string]string, r *http.Request) {

	// FIXME: make this configurable

	log.Printf("Debugging %v %v\n", r.Method, r.URL.Path)
	log.Println("Vars from gorilla/mux in the URL path:")
	for k, v := range vars {
		log.Printf("	%v => %v\n", k, v)
	}

	log.Println("Query string vars:")
	for k, v := range r.Form {
		log.Printf("	%v => %v\n", k, v)
	}

	log.Println("Post form vars:")
	for k, v := range r.PostForm {
		log.Printf("	%v => %v\n", k, v)
	}
}
