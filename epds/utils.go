// Package epds has some utils for accessing the files and managing the data in this repo
package epds

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/blgolden/igendec/params"

	"github.com/hjson/hjson-go"

	"github.com/blgolden/igendec/users"
)

var (
	filenameXref   = "comp_fn_pairs.hjson"
	filenameReadme = "README"
)

var (
	idField    = "ID"
	nameField  = "Name"
	scoreField = "Index"
	regField   = "RegNo"

	mainFields = []string{idField, nameField, regField}
)

// DatabasePath is the path to where the bull flat csv files are kept
var DatabasePath = "./epds/"

// ListDatabases returns a list of all the databases in this directory
func ListDatabases() []string {
	infos, err := ioutil.ReadDir(DatabasePath)
	if err != nil {
		return nil
	}
	databases := make([]string, 0, len(infos))
	for _, info := range infos {
		if info.IsDir() {
			databases = append(databases, info.Name())
		}
	}
	return databases
}

// Field describes a field in the CSV file read in from the comp_fn_pair.hjson file
type Field struct {
	Key     string `json:"name"`
	Header  string `json:"header"`
	Comment string `json:"comment"`
	Select  bool   `json:"select"`
	idx     int
}

// Bull holds the fields for a single bull
type Bull struct {
	Values []string
	Score  float64
}

// Fields returns a bull in string slice
func (b *Bull) Fields() []string {
	return append(b.Values, strconv.FormatFloat(b.Score, 'f', 2, 64))
}

// Database contains necessary fields for a bull database
type Database struct {
	Name         string
	Description  string
	databaseFile string
	Xref         map[string]Field
}

// NewDatabase returns an object with actionable methods on it
// for comparing jobs
func NewDatabase(name string) (*Database, error) {
	if info, err := os.Stat(filepath.Join(DatabasePath, name)); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("invalid name for a database")
	}
	db := &Database{Name: name}

	// Get the description
	if err := db.loadReadme(); err != nil {
		return nil, err
	}

	if err := db.loadXref(); err != nil {
		return nil, err
	}
	return db, nil
}

// FieldSlice returns a slice of the fields in the order that makes sense
func (db *Database) FieldSlice() []Field {
	slice := make([]Field, 0, len(db.Xref))
	added := make(map[string]bool)
	for _, name := range mainFields {
		if f, ok := db.Xref[name]; ok {
			slice = append(slice, f)
			added[name] = true
		}
	}
	traits := params.TraitKeys()
	for _, name := range traits {
		if f, ok := db.Xref[name]; ok {
			slice = append(slice, f)
			added[name] = true
		}
	}
	var subslice []Field
	for name, f := range db.Xref {
		if !added[name] {
			subslice = append(subslice, f)
		}
	}
	sort.Slice(subslice, func(i, j int) bool {
		return subslice[i].Key < subslice[j].Key
	})

	return append(slice, subslice...)
}

// TraitKeys returns the trait keys this database has fields for
func (db *Database) TraitKeys(set []string) []string {
	var keys []string
	s := make(map[string]bool, len(set))
	for _, id := range set {
		s[id] = true
	}
	for _, v := range params.TraitMap {
		_, ok1 := db.Xref[v.Short]
		_, ok2 := db.Xref[v.Display]
		if (ok1 || ok2) && (s[v.Short] || s[v.Display]) {
			keys = append(keys, v.Short)
		}
	}
	return keys
}

// CompareJob will take in a job and run it against the database
// and return a reader which has a formatted CSV, which has the jobs output
// run against it
func (db *Database) CompareJob(job *users.Job, fieldNames []string) (*bytes.Buffer, error) {

	// Find the filename for the databases CSV
	if err := db.findFilename(); err != nil {
		return nil, err
	}

	// Read in the CSV
	file, err := os.Open(filepath.Join(DatabasePath, db.Name, db.databaseFile))
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()
	r := csv.NewReader(file)
	r.ReuseRecord = true

	// Read header
	tokens, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("reading header")
	}

	// put into map
	header := make(map[string]int, len(tokens))
	for idx, field := range tokens {
		header[field] = idx
	}

	// helper function for mapping a field to a value in the file
	hmap := func(field string) (int, bool) {
		if key, ok := db.Xref[field]; ok {
			if idx, ok := header[key.Header]; ok {
				return idx, true
			}
		}
		return -1, false
	}

	// Create a convenience slice that will allow us to easily access and multiply the fields we need
	// to score a bull by
	type idxScaler struct {
		idx    int
		scaler float64
	}

	var components []idxScaler
	for _, ic := range job.Output {
		if idx, ok := hmap(ic.Key()); ok {
			components = append(components, idxScaler{idx, ic.MarginalEconomicValue})
		}
	}

	// Create a slice of ints so we can easily grab out the fields we want from each line
	var fields []int
	usedNames := make([]string, 0, len(fieldNames))
	for _, name := range fieldNames {
		if idx, ok := hmap(name); ok {
			fields = append(fields, idx)
			usedNames = append(usedNames, name)
		}
	}

	// read the rest of the file
	var bulls []Bull
	for {
		tokens, err = r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("reading file: %w", err)
		}

		bull := Bull{Values: make([]string, len(fields))}

		// Go through the components that we score a bull by and
		// build the score field
		for _, ic := range components {
			val, err := strconv.ParseFloat(tokens[ic.idx], 64)
			if err != nil {
				return nil, fmt.Errorf("converting field %s to float", tokens[ic.idx])
			}
			bull.Score += val * ic.scaler
		}

		// Iterate through the fields that the user wants and save them to the bull
		for i, idx := range fields {
			bull.Values[i] = tokens[idx]
		}

		bulls = append(bulls, bull)
	}

	// Sort
	sort.Slice(bulls, func(i, j int) bool {
		return bulls[i].Score > bulls[j].Score
	})

	// Write to buffer
	buf := &bytes.Buffer{}
	csvw := csv.NewWriter(buf)

	// Write header
	newHeader := append(usedNames, scoreField)
	if err = csvw.Write(newHeader); err != nil {
		return nil, fmt.Errorf("writing csv header: %w", err)
	}

	// Write bulls
	for _, bull := range bulls {
		if err := csvw.Write(bull.Fields()); err != nil {
			return nil, fmt.Errorf("writing csv: %w", err)
		}
	}

	csvw.Flush()

	return buf, nil
}

func (db *Database) loadXref() error {
	data, err := ioutil.ReadFile(filepath.Join(DatabasePath, db.Name, filenameXref))
	if err != nil {
		return fmt.Errorf("reading database xref: %w", err)
	}

	// Map each field we're searching for using hjson
	var m []interface{}
	if err = hjson.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("unmarshalling xref: %w", err)
	}

	// We can the remarshall/unmarshall the interface to automatically put the struct into our one
	xref := make(map[string]Field, len(m))
	for idx, v := range m {
		data, err = json.Marshal(v)
		if err != nil {
			return fmt.Errorf("remarshalling field %d: %w", idx, err)
		}
		f := Field{}
		if err = json.Unmarshal(data, &f); err != nil {
			return fmt.Errorf("unmarshalling field %d: %w", idx, err)
		}
		f.idx = idx
		xref[f.Key] = f
	}

	// Check we have the fields we require
	for _, f := range []string{idField} {
		if _, ok := xref[f]; !ok {
			return fmt.Errorf("no %s field in database", f)
		}
	}

	db.Xref = xref
	return nil
}

func (db *Database) loadReadme() error {
	data, err := ioutil.ReadFile(filepath.Join(DatabasePath, db.Name, filenameReadme))
	if err != nil {
		return fmt.Errorf("reading database readme: %w", err)
	}
	db.Description = string(data)
	return nil
}

func (db *Database) findFilename() error {
	// Find the database filename
	infos, err := ioutil.ReadDir(filepath.Join(DatabasePath, db.Name))
	if err != nil {
		return fmt.Errorf("reading database dir: %w", err)
	}
	for _, info := range infos {
		if !info.IsDir() && filepath.Ext(info.Name()) == ".csv" {
			if db.databaseFile != "" {
				return fmt.Errorf("multiple csv files in directory")
			}
			db.databaseFile = info.Name()
		}
		//fmt.Println(err)
	}
	if db.databaseFile == "" {
		return fmt.Errorf("could not find a csv database")
	}
	return nil
}

// Test allows you to verify the database can access the files it needs to
// does not validify the CSV, just checks it exists
func (db *Database) Test() error {
	// Load in the files
	if err := db.loadXref(); err != nil {
		return fmt.Errorf("failed to parse xref: %w", err)
	}
	if err := db.findFilename(); err != nil {
		return fmt.Errorf("database has no .csv file")
	}
	file, err := os.Open(filepath.Join(DatabasePath, db.Name, db.databaseFile))
	if err != nil {
		return fmt.Errorf("coud not open file %s: %w", filepath.Join(DatabasePath, db.Name, db.databaseFile), err)
	}
	return file.Close()
}
