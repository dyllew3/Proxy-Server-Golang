package main
import(
	"bytes"
	"bufio"
	"fmt"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"
)
const dumpBody = true

type Cache struct {

	// Use sync.Map as it a thread safe Dictionary
	// Maps url to byte array
	Elems sync.Map
}

func CreateCache() *Cache {

	return &Cache{}
}

// Check if cache contains the response to the http request
// If in cache byte array and true are returned if not in cache
// nil and false are returned
func Hit(req *http.Request,cache *Cache) ([]byte,bool) {

	elems := cache.Elems
	val, ok := elems.Load(req.URL.String())
	if ok {
		return val.([]byte),true
	}
	return nil,false
}

func Expired(req *http.Request,cache *Cache) bool{

	elems := cache.Elems
	val, ok := elems.Load(req.URL.String())
	r := bufio.NewReader(bytes.NewReader(val.([]byte)))
	resp, _ := http.ReadResponse(r, req)

	arr, val_key := resp.Header["Expires"]
	// Evaluate whether it has expires field and it is present in the cache
	if ok && val_key {
		time_val, _ := time.Parse(time.RFC1123, arr[0])
		if int64(time_val.Sub(time.Now())) < 0 {
			return true
		}
	}
	return false

}


func Insert(req *http.Request, resp * http.Response, cache *Cache){


	body, err  := httputil.DumpResponse(resp, dumpBody)
	if err != nil {
		fmt.Println("Unable to add to cache")
	} else {
		req_str := req.URL.String()
		cache.Elems.Store(req_str, body)
	}
}
