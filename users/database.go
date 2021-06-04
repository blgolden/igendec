package users

// Defines the interface for the user database

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/blgolden/igendec/logger"
	"github.com/blgolden/igendec/params"
)

var UsersPath string

// Init sets up the packages database
// If you want to change the type
func Init() {
	database = NewLocalDatabase(UsersPath)
}

var database *LocalDatabase

// Prefixes for file structure of iGenDec
const (
	PrefixUsers         = "users/"
	PrefixJobs          = "jobs/"
	FileProfileFilename = "profile.hjson"
	FileMasterFilename  = "masterParams.hjson"
	FileEcoFilename     = "ecoParams.hjson"

	//FileJobOutput         = "output.hjson"
	FileJobOutput         = "output.json"
	FileJobProcessingFlag = ".processing"
)

// Database errors
var (
	ErrUserExists      = errors.New("user exists")
	ErrUserDoesntExist = errors.New("user does not exist")
)

// PathToJobFile returns the path to a job file for a user
// If you want the general path, pass in empty string for jobfile parameter
func PathToJobFile(user, job, file string) string {
	return filepath.Join(PathToJobsDir(user), job, file)
}

// PathToJobsDir returns the path to the directory where we keep all of the jobs
func PathToJobsDir(user string) string {
	return filepath.Join(PathToUserDir(user), PrefixJobs)
}

// PathToUserDir returns the path to a directory for a user
func PathToUserDir(user string) string {
	return filepath.Join(database.root, PrefixUsers, user)
}

// PathToUserFile returns the path to a general file for a user
func PathToUserFile(user, file string) string {
	return filepath.Join(PathToUserDir(user), file)
}

// Database implementation

// LocalDatabase structure is an implementation of Database interface for keeping files in local tree
type LocalDatabase struct {
	root    string      // Should be absolute
	perm    os.FileMode // permission to write files out with
	dirperm os.FileMode // permission to create directories with
}

// NewLocalDatabase returns a new local implementation of Database
func NewLocalDatabase(root string) *LocalDatabase {
	if !filepath.IsAbs(root) {
		if newpath, err := filepath.Abs(root); err != nil {
			logger.Fatal("creating local database: converting path '%s' to an absolute path: %s", root, err)
		} else {
			root = newpath
		}
	}
	root = strings.TrimRight(root, "/") + "/"
	os.MkdirAll(root, 0755)
	return &LocalDatabase{root, 0644, 0755}
}

// Get returns a user
func (db *LocalDatabase) Get(username string) (*User, error) {
	data, err := ioutil.ReadFile(PathToUserFile(username, FileProfileFilename))
	if err != nil {
		return nil, ErrUserDoesntExist
	}
	return NewUserFromBytes(data)
}

// Create makes a new user
func (db *LocalDatabase) Create(user *User) error {
	if db.exists(user.Username) {
		return ErrUserExists
	}

	data, err := user.Bytes()
	if err != nil {
		return err
	}
	// Make directory structure for user
	os.MkdirAll(PathToUserDir(user.Username), db.dirperm)
	if err = os.Mkdir(PathToJobsDir(user.Username), db.dirperm); err != nil {
		return err
	}
	// return user
	return ioutil.WriteFile(PathToUserFile(user.Username, FileProfileFilename), data, db.perm)
}

// Update takes in the details of a user and overwrites the current profile entry
func (db *LocalDatabase) Update(user *User) error {
	if !db.exists(user.Username) {
		return ErrUserExists
	}

	data, err := user.Bytes()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(PathToUserFile(user.Username, FileProfileFilename), data, db.perm)
}

// GetIndexParams returns indexParams.hjson file for the user
func (db *LocalDatabase) GetIndexParams(user string) (*params.MasterParams, error) {
	return params.MasterParamsFromFile(PathToUserFile(user, FileMasterFilename))
}

// SetMasterParams writes the given index params to the database
func (db *LocalDatabase) SetMasterParams(user string, ip *params.MasterParams) error {
	data, err := ip.Bytes()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(PathToUserFile(user, FileMasterFilename), data, db.perm)
}

// GetEcoParams returns ecoParams.hjson file for the user
func (db *LocalDatabase) GetEcoParams(user string) (*params.EcoParams, error) {
	return params.EcoParamsFromFile(PathToUserFile(user, FileEcoFilename))
}

// SetEcoParams writes the given eco params to the database
func (db *LocalDatabase) SetEcoParams(user string, ep *params.EcoParams) error {
	data, err := ep.Bytes()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(PathToUserFile(user, FileEcoFilename), data, db.perm)
}

// ListJobs returns a list of the jobs a user has
func (db *LocalDatabase) ListJobs(user string) []string {
	filelist, err := ioutil.ReadDir(PathToJobsDir(user))
	if err != nil {
		return nil
	}
	var jobs = make([]string, len(filelist))
	for idx, info := range filelist {
		jobs[idx] = info.Name()
	}
	return jobs
}

// GetJobFilename returns the path to the given job file
func (db *LocalDatabase) GetJobFilename(user, job, file string) string {
	p := PathToJobFile(user, job, file)
	os.MkdirAll(filepath.Dir(p), database.dirperm)
	return p
}

// DeleteJob will remove the given job if it exists
func (db *LocalDatabase) DeleteJob(user, job string) error {
	return os.RemoveAll(PathToJobFile(user, job, ""))
}

// Returns true if user exists - or is reachable, false otherwise
func (db *LocalDatabase) exists(username string) bool {
	_, err := os.Stat(PathToUserFile(username, FileProfileFilename))
	return err == nil
}
