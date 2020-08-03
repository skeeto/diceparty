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

func generateRoll(who string) roll {
	mutex.Lock()
	defer mutex.Unlock()
	roll := roll{
		ID:     len(rolls),
		When:   time.Now(),
		Who:    who,
		Result: roller.Roll(),
	}
	rolls = append(rolls, roll)
	update.Broadcast()
	return roll
}

func handleRoll(w http.ResponseWriter, r *http.Request) {
	who, err := url.PathUnescape(r.URL.RawQuery)
	if err != nil {
		who = r.URL.RawQuery
	}
	roll := generateRoll(who)
	log.Printf("ROLL %s %s %+d", r.RemoteAddr, who, roll.Result)
}

func pollResponse(begin int) []byte {
	mutex.Lock()
	defer mutex.Unlock()
	for begin >= len(rolls) {
		update.Wait()
	}
	buf, _ := json.Marshal(rolls[begin:])
	return buf
}

func handlePoll(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	begin64, err := strconv.ParseUint(r.URL.RawQuery, 10, 32)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	begin := int(begin64)

	log.Println("POLL", r.RemoteAddr, begin)
	buf := pollResponse(begin)
	w.Write(buf)
}

func handleHTML(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		log.Println("/", r.RemoteAddr)
		w.Write(pageHTML)
	} else {
		w.WriteHeader(404)
	}
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
