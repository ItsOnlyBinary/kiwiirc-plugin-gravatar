package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"github.com/kiwiirc/webircgateway/pkg/webircgateway"

	// Database drivers
	_ "github.com/go-sql-driver/mysql"
)

// Config the structure of our json
type Config struct {
	DSN                   string `json:"dsn"`
	Query                 string `json:"query"`
	Salt                  string `json:"salt"`
	GravatarURL           string `json:"gravatar_url"`
	CacheLife             string `json:"cache_life"`
	CacheInterval         string `json:"cache_interval"`
	cacheLifeDuration     time.Duration
	cacheIntervalDuration time.Duration

	// Extra options for running standalone
	ListenAddress      string   `json:"listen_addr"`
	AllowedOrigins     []string `json:"allow_origins"`
	AllowedOriginsGlob []glob.Glob
}

// Default config
var config = &Config{
	DSN:            "root:@tcp(127.0.0.1:3306)/anope",
	Query:          "SELECT display AS `account`, email FROM anope_db_NickCore WHERE display = ? LIMIT 0,1;",
	GravatarURL:    "//www.gravatar.com/avatar",
	CacheLife:      "6h",
	CacheInterval:  "15m",
	AllowedOrigins: []string{"*"},
}

// Account represents cached data
type Account struct {
	Account  string `json:"account"`
	Gravatar string `json:"gravatar"`
	email    string
	cacheAge time.Time
}

var gateway *webircgateway.Gateway
var db *sql.DB

var cache map[string]*Account
var cacheLock sync.RWMutex

// Start is called by webircgateway
func Start(g *webircgateway.Gateway, pluginsQuit *sync.WaitGroup) {
	gateway = g
	logError(1, "Starting")

	startGravatar(gateway.HttpRouter)
}

func startGravatar(httpRouter *http.ServeMux) {
	configFile := flag.String("gravatar-config", "gravatar.config.json", "Config file location")
	flag.Parse()

	// Attempt to load our config file
	if loadConfig(*configFile) == nil {
		return
	}

	// Make sure the user provided a salt
	if config.Salt == "" {
		logError(3, "Config missing salt")
		return
	}

	cache = make(map[string]*Account)
	go cacheCleanup()
	dbConnect()
	httpRouter.Handle("/gravatar/", http.HandlerFunc(handleGravatar))
}

func handleGravatar(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	// Check for allowed Origin
	originHeader := strings.ToLower(req.Header.Get("Origin"))
	if !isOriginAllowed(originHeader) {
		hookError(w, http.StatusForbidden)
		return
	}

	// Check for allowed Method
	if req.Method != "GET" {
		hookError(w, http.StatusMethodNotAllowed)
		return
	}

	urlPath, rawAccount := path.Split(req.URL.Path)
	if urlPath != "/gravatar/" || rawAccount == "" {
		hookError(w, http.StatusBadRequest)
		return
	}

	lcAccount := strings.ToLower(rawAccount)

	// Check cache
	cacheAccount := false
	cacheLock.RLock()
	account := cache[lcAccount]
	cacheLock.RUnlock()

	// Account is not cached or timedout so attempt to retrive from db
	if account == nil || time.Since(account.cacheAge) > config.cacheLifeDuration {
		account = &Account{}
		query := db.QueryRow(config.Query, lcAccount)
		err := query.Scan(&account.Account, &account.email)
		if err == sql.ErrNoRows {
			// if no row was found use account name plus salt
			// this is to prevent fishing for valid accounts
			account.email = lcAccount + config.Salt
		} else if err != nil {
			logError(3, "Database query error: "+err.Error())
			hookError(w, http.StatusInternalServerError)
			return
		}

		account.Gravatar = getHash(account.email)
		account.cacheAge = time.Now()
		cacheAccount = true
	}

	gravURL, err := url.Parse(config.GravatarURL)
	if err != nil {
		hookError(w, http.StatusInternalServerError)
		return
	}

	gravURL.RawQuery = req.URL.RawQuery
	gravURL.Path = path.Join(gravURL.Path, account.Gravatar)

	w.Header().Set("Access-Control-Allow-Origin", originHeader)
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age:%.0f", config.cacheLifeDuration.Seconds()))

	http.Redirect(w, req, gravURL.String(), 302)

	if cacheAccount {
		cacheLock.Lock()
		cache[lcAccount] = account
		cacheLock.Unlock()
	}
}

func dbConnect() error {
	var err error
	db, err = sql.Open("mysql", config.DSN)
	if err != nil {
		logError(3, "Database connect error: "+err.Error())
	}
	err = db.Ping()
	if err != nil {
		logError(3, "Database ping error: "+err.Error())
	}
	return err
}

func getHash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func cacheCleanup() {
	for {
		time.Sleep(config.cacheIntervalDuration)
		remove := make([]string, 0)
		cacheLock.RLock()
		logError(1, "Cleaning cache (items: %v)", len(cache))
		for lcAccount, account := range cache {
			if time.Since(account.cacheAge) > config.cacheLifeDuration {
				remove = append(remove, lcAccount)
			}
		}
		cacheLock.RUnlock()

		if len(remove) > 0 {
			removeCount := 0
			cacheLock.Lock()
			for _, lcAccount := range remove {
				if time.Since(cache[lcAccount].cacheAge) > config.cacheLifeDuration {
					delete(cache, lcAccount)
					removeCount++
				}
			}
			cacheLock.Unlock()
			logError(1, "Removed from cache: %v", removeCount)
		}
	}
}

func loadConfig(configFile string) *Config {
	raw, err := ioutil.ReadFile(configFile)
	if err != nil {
		logError(3, "Config read error: "+err.Error())
		return nil
	}
	err = json.Unmarshal(raw, &config)
	if err != nil {
		logError(3, "Config unmarshal error: "+err.Error())
		return nil
	}
	logError(1, "Config loaded: "+configFile)

	if config.cacheLifeDuration, err = time.ParseDuration(config.CacheLife); err != nil {
		logError(3, "Config cache_life parse error: "+err.Error())
		return nil
	}

	if config.cacheIntervalDuration, err = time.ParseDuration(config.CacheInterval); err != nil {
		logError(3, "Config cache_interval parse error: "+err.Error())
		return nil
	}

	config.AllowedOriginsGlob = []glob.Glob{}
	for i := 0; i < len(config.AllowedOrigins); i++ {
		newAllowedOrigin, err := glob.Compile(config.AllowedOrigins[i])
		if err != nil {
			logError(3, "Config allow_origin failed to parse glob: "+err.Error())
			continue
		}
		config.AllowedOriginsGlob = append(config.AllowedOriginsGlob, newAllowedOrigin)
	}

	return config
}

func isOriginAllowed(originHeader string) bool {
	if gateway != nil {
		return gateway.IsClientOriginAllowed(originHeader)
	}

	// Empty list of origins = all origins allowed
	if len(config.AllowedOrigins) == 0 {
		return true
	}

	// No origin header = running on the same page
	if originHeader == "" {
		return true
	}

	foundMatch := false

	for _, originMatch := range config.AllowedOriginsGlob {
		if originMatch.Match(originHeader) {
			foundMatch = true
			break
		}
	}
	return foundMatch
}

func hookError(w http.ResponseWriter, i int) {
	http.Error(w, fmt.Sprintf("%d %s", i, http.StatusText(i)), i)
}

func logError(level int, message string, args ...interface{}) {
	if gateway != nil {
		gateway.Log(level, fmt.Sprintf("[plugin-gravatar] "+message, args...))
	} else {
		log.Printf(message+"\n", args...)
	}
}
