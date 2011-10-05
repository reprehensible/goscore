package goscore

import (
	"fmt"
	"http"
	"appengine"
	"appengine/user"
	"appengine/datastore"
	"template"
	"time"
	"strconv"
)

type Game struct {
	P1 string
	P1score int
	P2 string
	P2score int
	Played datastore.Time
}

var (
	appTemplate = template.MustParseFile("app.html", nil)
)

func init() {
	http.HandleFunc("/games", games)
	http.HandleFunc("/", root)
}

func initProtectedPage(w http.ResponseWriter, r *http.Request) (c appengine.Context, u *user.User, die bool) {
	die = false

	c = appengine.NewContext(r)

	u = user.Current(c)

	if u == nil {
		die = true
		url, err := user.LoginURL(c, r.URL.String())
		if err != nil {
			http.Error(w, err.String(), http.StatusInternalServerError)
		} else {
			w.Header().Set("Location", url)
			w.WriteHeader(http.StatusFound)
		}
	}

	return c, u, die //hah, unintentional
}

func root(w http.ResponseWriter, r *http.Request) {
	c, _, die := initProtectedPage(w, r)
	if die { return }

	q := datastore.NewQuery("Game").Order("-Played").Limit(10)
	games := make([]Game, 0, 10)
	if _, err := q.GetAll(c, &games); err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
	}

	err := appTemplate.Execute(w, games)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
	}
}

func games(w http.ResponseWriter, r *http.Request) {
	c, u, die := initProtectedPage(w, r)
	if die { return }


	switch r.Method {
	case "POST":
		gamesPost(w, r, u, c)
	case "GET":
		http.Error(w, "Method Not Implemented", http.StatusInternalServerError)
		gamesGet(w, r, u, c)
	default:
		http.Error(w, "Method Not Implemented", http.StatusInternalServerError)
	}
}

func gamesPost(w http.ResponseWriter, r *http.Request, u *user.User, c appengine.Context) {

	game := Game{
		P1: r.FormValue("player1"),
		P2: r.FormValue("player2"),
		Played: datastore.SecondsToTime(time.Seconds()),
	}
	scorefields :=  map[string]int{
		"Player1score": 0,
		"Player2score": 0,
	}
	for k, _ := range scorefields{
		score, err := strconv.Atoi(r.FormValue(k))
		if err != nil {
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}
		scorefields[k] = score
	}
	game.P1score = scorefields["Player1score"]
	game.P2score = scorefields["Player2score"]

	_, err := datastore.Put(c, datastore.NewIncompleteKey("Game"), &game)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	rurl := fmt.Sprintf("http://%v/", r.Host)
	w.Header().Set("Location", rurl)
	w.WriteHeader(http.StatusFound)
}

func gamesGet(w http.ResponseWriter, r *http.Request, u *user.User, c appengine.Context) {
	err := appTemplate.Execute(w, nil)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
	}
}
