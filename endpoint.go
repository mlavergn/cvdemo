package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// City data
type City struct {
	name      string
	longitude float64
	latitude  float64
	country   string
	region    string
}

// QueryResultEntry data
type QueryResultEntry struct {
	Name      string  `json:"name"`
	Latitude  string  `json:"latitude"`
	Longitude string  `json:"longitude"`
	Score     float64 `json:"score"`
}

// QueryResult JSON input
type QueryResult struct {
	Suggestions []*QueryResultEntry `json:"suggestions"`
}

func newQueryResult() *QueryResult {
	return &QueryResult{Suggestions: make([]*QueryResultEntry, 0)}
}

func (result *QueryResult) add(name string, latitude float64, longitude float64, score float64) {
	result.Suggestions = append(result.Suggestions, &QueryResultEntry{
		Name:      name,
		Latitude:  strconv.FormatFloat(latitude, 'f', 5, 64),
		Longitude: strconv.FormatFloat(longitude, 'f', 5, 64),
		Score:     score,
	})
}

// NewCity helper
func NewCity(name string, latitude float64, longitude float64, country string, region string) *City {
	return &City{
		name:      name,
		longitude: longitude,
		latitude:  latitude,
		country:   country,
		region:    region,
	}
}

// TrieNodeMap typedef for map of trie nodes
type TrieNodeMap map[rune]*TrieNode

// TrieNode is the node definition for the trie
type TrieNode struct {
	char  rune
	ids   []int64
	nodes TrieNodeMap
}

// NewTrieNode helper
func NewTrieNode(char rune, cid int64) *TrieNode {
	return &TrieNode{
		char:  char,
		ids:   []int64{cid},
		nodes: make(TrieNodeMap),
	}
}

func (trieNode *TrieNode) addCid(cid int64) {
	trieNode.ids = append(trieNode.ids, cid)
}

// Trie is the main data structure
type Trie struct {
	cities map[int64]*City
	nodes  TrieNodeMap
}

// NewTrie helper
func NewTrie() *Trie {
	return &Trie{
		cities: make(map[int64]*City),
		nodes:  make(map[rune]*TrieNode),
	}
}

func (trie *Trie) find(prefix string, latitude float64, longitude float64) *QueryResult {
	result := newQueryResult()

	var node *TrieNode
	// walk the trie until exhaustion of the prefix
	for _, c := range strings.ToLower(prefix) {
		// skip [a-z] query characters
		if c < 97 || c > 122 {
			continue
		}

		if node == nil {
			// assign an initial node
			node = trie.nodes[c]
		} else {
			node = node.nodes[c]
		}
		if node == nil {
			// not found
			return result
		}
	}

	// failed to find trie value
	if node == nil {
		println("Failed to find trie entry for", prefix, latitude, longitude)
		return result
	}

	// process the CIDs
	for _, cid := range node.ids {
		city := trie.cities[cid]
		if city != nil {
			fullName := fmt.Sprintf("%s, %s, %s", city.name, city.region, city.country)
			score := trie.score(cid, prefix, latitude, longitude)
			result.add(fullName, city.latitude, city.longitude, score)
		}
	}

	// sort the results in descending order of score
	sort.Slice(result.Suggestions, func(i, j int) bool {
		return result.Suggestions[i].Score > result.Suggestions[j].Score
	})

	return result
}

func (trie *Trie) score(cid int64, query string, latitude float64, longitude float64) float64 {
	city := trie.cities[cid]

	divisor := math.Max(float64(len(query)), float64(len(city.name))) / math.Min(float64(len(query)), float64(len(city.name)))
	prefField := math.Min(40.0/divisor, 40.0)

	divisor = math.Abs(latitude - city.latitude)
	latField := math.Min(30.0/divisor, 30.0)

	divisor = math.Abs(longitude - city.longitude)
	longField := math.Min(30.0/divisor, 30.0)

	return math.Round((prefField+latField+longField)/100.0*10) / 10
}

var extendedLookup = map[string]bool{"-": true, "'": true, ".": true, ",": true, " ": true, "(": true, ")": true, "1": true}

func (trie *Trie) add(cid int64, name string, asciiName string, latitude float64, longitude float64, country string, region string) bool {
	city := NewCity(name, latitude, longitude, country, region)
	trie.cities[cid] = city

	var node *TrieNode
	var child *TrieNode

	for i, c := range strings.ToLower(asciiName) {
		// is c within [a-z]
		if c < 97 || c > 122 {
			if extendedLookup[string(c)] == false {
				println("Unknown character [%s] at [%i] in [%s]", c, i, name)
			}
			continue
		}
		if node == nil {
			child = trie.nodes[c]
		} else {
			child = node.nodes[c]
		}
		if child == nil {
			child = NewTrieNode(c, cid)
			if node == nil {
				trie.nodes[c] = child
			} else {
				node.nodes[c] = child
			}
		} else {
			child.addCid(cid)
		}
		node = child
	}

	return true
}

var regionLookup = map[string]string{
	"01": "AB",
	"02": "BC",
	"03": "MB",
	"04": "NB",
	"05": "NL",
	"07": "NS",
	"08": "ON",
	"09": "PE",
	"10": "QC",
	"11": "SK",
	"12": "YT",
	"13": "NT",
	"14": "NU",
}

// Load the tsv file and parse into a trie
func (trie *Trie) load(path string) bool {
	// file, err := os.Open(path)
	// defer file.Close()
	// if err != nil {
	// 	log.Fatal(err)
	// 	return false
	// }
	// scanner := bufio.NewScanner(file)
	// for scanner.Scan() {
	// 	fmt.Println(scanner.Text())
	// }

	timer := time.Now()
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// fields: 0: id | 1: name | 2: ascii | 3: altname | 4: lat | 5: long | 8: country | 10: region
	var lines = strings.Split(string(bytes), "\n")
	for i, line := range lines {
		fields := strings.Split(line, "\t")
		if len(fields) < 10 || i == 0 {
			println("Skipping line:", i)
			continue
		}
		var region string
		if fields[8] == "US" {
			region = fields[10]
		} else {
			region = regionLookup[fields[10]]
			if region == "" {
				println("Failed region code lookup:", fields)
				continue
			}
		}
		var country string
		if fields[8] == "US" {
			country = "USA"
		} else {
			country = "Canada"
		}
		cid, err0 := strconv.ParseInt(fields[0], 10, 64)
		latitude, err1 := strconv.ParseFloat(fields[4], 64)
		longitude, err2 := strconv.ParseFloat(fields[5], 64)
		if err0 == nil {
			if err1 != nil {
				latitude = 0.00001
			}
			if err2 != nil {
				longitude = 0.00001
			}
			trie.add(cid, fields[1], fields[2], latitude, longitude, country, region)
		}
	}

	println("Trie built in", time.Since(timer)/time.Millisecond, "ms")

	return true
}

func (trie *Trie) defaultHandler(resp http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	q := req.FormValue("q")

	// strip extended characters
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	prefix, _, _ := transform.String(t, q)

	latitude, errLat := strconv.ParseFloat(req.FormValue("latitude"), 64)
	longitude, errLong := strconv.ParseFloat(req.FormValue("longitude"), 64)

	if errLat != nil {
		latitude = 0.00001
	}
	if errLong != nil {
		longitude = 0.00001
	}

	timer := time.Now()
	result := trie.find(prefix, latitude, longitude)
	println("Search for", prefix, "completed in", time.Since(timer)/time.Microsecond, "μs")

	b, err := json.Marshal(result)
	if err != nil {
		println("JSON serialization error", err)
	}

	resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	resp.Write([]byte(b))
}

func main() {
	trie := NewTrie()
	trie.load("cities_canada-usa.tsv")
	http.HandleFunc("/", trie.defaultHandler)
	println("Serving on port 8080")
	http.ListenAndServe(":8080", nil)
}
