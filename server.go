package main


import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const JsonFile = "./blocked.json"
const ConsoleAdr = "://console"


type Blocked struct{

	URL string `json:url`

}

var blockedUrls []Blocked
var cachedUrls *Cache


// converts a Blocked struct to a string
func (b Blocked) toString() string {
	bytes, err := json.Marshal(b)
  if err != nil {
  	fmt.Println(err.Error())
    os.Exit(1)
  }

  return string(bytes)

}


//loads the blocked urls from the json file
func LoadBlocked() []Blocked {
	file, err := ioutil.ReadFile(JsonFile)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
    os.Exit(1)
	}
	var result []Blocked
	json.Unmarshal(file, &result)
	return result
}


// Write the blocklist back to the file
func WriteBlocked() {

	result := []byte("[")
	length := len(blockedUrls)
	// Convert the urls into a json format
	for index,item := range blockedUrls{
		bytes, err := json.Marshal(item)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		result = append(result, bytes...)
		if index < length - 1 {
					result =  append(result, []byte(",")...)
		}

	}
	result = append(result, []byte("]")...)
	ioutil.WriteFile("./blocked.json",result, 0644)
}


// This is repsonsible for dealing with HTTP headers
// Checks destination then sends the request to its destination
func HttpHeader(w http.ResponseWriter, req *http.Request){
	client := &http.Client{}
	fmt.Printf("HTTP request received for %s\n", req.URL.String())
	req.RequestURI = ""
	// If request is for console call Console handler
	if strings.Contains(req.URL.String(), ConsoleAdr) {
		Console(w, req)

	} else if IsBlocked(req.URL.String()) {
		fmt.Fprintf(w, "%s", "Blocked by proxy")

	} else {
		var resp *http.Response
		// check if in cache
		val, hit := Hit(req, cachedUrls)
		if !hit || Expired(req, cachedUrls) {
			// If not in cache or expired insert packet into cache
			resp , _ = client.Do(req)
			fmt.Println("Miss")
			Insert(req, resp, cachedUrls)
		} else {
			fmt.Println("Hit")
			r := bufio.NewReader(bytes.NewReader(val))
			resp, _ = http.ReadResponse(r, req)

		}
	  a := w.Header()
	  b := resp.Header
	  FormatHeader(a, b)
	  w.WriteHeader(resp.StatusCode)
		// Copy body to ResponseWriter
		io.Copy(w, resp.Body)
	}

}

// Checks if the url is in the blocked list
func IsBlocked(url string) bool  {
	for _,blocked := range blockedUrls {
		if strings.Contains(url, blocked.URL) {
			return true
		}

	}
	return false
}

// This handles the console page which adds and removes
// urls from the blocklist
func Console(w http.ResponseWriter, req *http.Request){
	if strings.Contains(req.URL.String(), "/blocked") {
		// create string of blocked urls which will be displayed to user
		blockString := ""
		for _, item := range blockedUrls {
			blockString = fmt.Sprintf("%s <li>%s</li>", blockString, item.URL)
		}
		fmt.Fprintf(w,"<html> <h1>Block list</h1> <body><ul> %s </ul> </body> </html>", blockString)
	} else if strings.Contains(req.URL.String(), "/remove") &&   req.Method == "POST"  {
		// This iterates through the blocklist and removes the specified url from the list
		url := req.PostFormValue("remove_url")
		for index, item := range blockedUrls {
			if strings.Compare(url, item.URL) == 0 {
				blockedUrls[len(blockedUrls)-1], blockedUrls[index] = blockedUrls[index], blockedUrls[len(blockedUrls)-1]
				blockedUrls = blockedUrls[:len(blockedUrls)-1]
				break
			}
		}
		defer WriteBlocked()
		fmt.Fprintf(w,"successfully removed %s from the blocked list", url)

	} else if req.Method == "POST" {
		// Get the specified url to be added to the blocked list
		url := req.PostFormValue("url")
		newUrl := Blocked{ URL: url,}
		blockedUrls = append(blockedUrls, newUrl)
		defer WriteBlocked()
		// Print new blocked list for user
		blockString := ""
		for _, item := range blockedUrls {
			blockString = fmt.Sprintf("%s <li>%s</li>", blockString, item.URL)

		}
		fmt.Fprintf(w,"<html>	<h1>New blocked list</h1><body><p>%s added</p><p>List currently has the following urls</p><ul> %s </ul></body></html>", url, blockString)


	} else {

		http.ServeFile(w, req, "base.html")
	}
}


// Formats header so it can be used by client
func FormatHeader(dest , src http.Header ){

	for k, vs := range src {
		for _, v := range vs {
			dest.Add(k, v)
		}
	}
	dest.Del("Proxy-Connection")
	dest.Del("Proxy-Authenticate")
	dest.Del("Proxy-Authorization")
	dest.Del("Connection")
}

// This forwards the bytes from one tcp connection to another
func CopyTo(dest, src net.Conn){
	defer src.Close()
	io.Copy(dest, src)
	
}


func HttpsHeader(w http.ResponseWriter, req *http.Request){
	if !IsBlocked(req.URL.String()) {
		// Establish tcp connection with destination
		dest_conn, err := net.Dial("tcp", req.URL.Host)
		hjk, works := w.(http.Hijacker)
		if !works {
			http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
			return
		}
		client_conn, _, err := hjk.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		// Prints https request
		fmt.Printf("HTTPS request received for %s\n", req.URL.String())
		// accepts the https upgrade
		client_conn.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))

		// Now all thats left is to forward the https requests and bytes
		// from the client to the destination and the responses back to the
		// client
		go CopyTo(dest_conn, client_conn)
		go CopyTo(client_conn, dest_conn)
	}
}

// Handler function for server which determines which function
// to use based on request method
func Handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	if r.Method != http.MethodConnect {
		HttpHeader(w, r)
	}	else {
		HttpsHeader(w,r)

	}
	fmt.Println("Time taken to serve is " + time.Since(start).String())
}




func main(){
	blockedUrls = LoadBlocked()
	cachedUrls =  CreateCache()
	server := http.Server{
    Addr: ":8080",
    Handler: http.HandlerFunc(Handler),
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalln("Error: %v", err)
	}

}
