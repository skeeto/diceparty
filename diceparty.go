package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

type roll struct {
	ID     int       `json:"id"`
	When   time.Time `json:"when"`
	Who    string    `json:"who"`
	Result int       `json:"result"`
}

var mutex sync.Mutex
var rolls []roll
var roller = newDice()
var update = sync.NewCond(&mutex)
var pageHTML []byte

func handleRoll(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	who, err := url.PathUnescape(r.URL.RawQuery)
	if err != nil {
		who = r.URL.RawQuery
	}
	rolls = append(rolls, roll{
		ID:     len(rolls),
		When:   time.Now(),
		Who:    who,
		Result: roller.Roll(),
	})
	update.Broadcast()
}

func handlePoll(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	begin64, err := strconv.ParseUint(r.URL.RawQuery, 10, 32)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	begin := int(begin64)

	mutex.Lock()
	defer mutex.Unlock()
	for begin >= len(rolls) {
		update.Wait()
	}
	buf, _ := json.Marshal(rolls[begin:])
	w.Write(buf)
}

func handleHTML(w http.ResponseWriter, r *http.Request) {
	w.Write(pageHTML)
}

func loadHTML() error {
	f, err := os.Open("index.html")
	if err != nil {
		return err
	}
	defer f.Close()

	html, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	pageHTML = html
	return nil
}

func main() {
	if err := loadHTML(); err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", handleHTML)
	http.HandleFunc("/roll", handleRoll)
	http.HandleFunc("/poll", handlePoll)
	http.ListenAndServe(":8080", nil)
}
