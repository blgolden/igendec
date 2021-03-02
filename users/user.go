package users

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/blgolden/igendec/params"
	"golang.org/x/crypto/bcrypt"
)

// Password variables
const (
	bcryptCost        = 7
	PasswordMinLength = 8
	PasswordMaxLength = 64
)

// Password errors
var (
	ErrInvalidPassword   = errors.New("Invalid password")
	ErrIncorrectPassword = errors.New("Incorrect password")
	ErrUserHasNoPassword = errors.New("User does not have a password")
)

// User has the fields a user needs
type User struct {
	Firstname string
	Surname   string

	Email string

	Location string

	Username string

	// Password - hashed and salted
	Password []byte
}

// NewUser returns a new user with only the Username field filled in
// Use 'user.NewUser(username).Get()' to get a user from the database
func NewUser(username string) *User {
	return &User{
		Username: username,
	}
}

// NewUserFromBytes returns a new user by parsing the bytes with JSON
func NewUserFromBytes(data []byte) (*User, error) {
	user := &User{}
	err := json.Unmarshal(data, user)
	return user, err
}

// ToMap returns the values we need from the struct in a fiber compatible map
func (u *User) ToMap(m map[string]interface{}) map[string]interface{} {
	m["Firstname"] = u.Firstname
	m["Surname"] = u.Surname
	m["Email"] = u.Email
	m["Location"] = u.Location
	m["Username"] = u.Username
	return m
}

// Get a user from the database. Will populate the callers fields
func (u *User) Get() (*User, error) {
	new, err := database.Get(u.Username)
	if err != nil {
		return nil, err
	}
	u.Firstname = new.Firstname
	u.Surname = new.Surname
	u.Email = new.Email
	u.Location = new.Location
	u.Password = new.Password
	return u, nil
}

// Save the user to the database
// This is a create operation, to update use user.Update()
func (u *User) Save() error {
	return database.Create(u)
}

// Update the users details in the database
// This will fail if the user doesn't exist
func (u *User) Update() error {
	return database.Update(u)
}

// Exists returns true if the user exists
func (u *User) Exists() bool {
	_, err := database.Get(u.Username)
	return err == nil
}

// ValidateAndHashPassword validates and hashes the provided password and saves it to the user struct
func (u *User) ValidateAndHashPassword(password string) (err error) {
	if len(password) < PasswordMinLength || PasswordMaxLength < len(password) {
		return ErrInvalidPassword
	}
	u.Password, err = bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return
}

// ComparePassword returns true if the password is this users, false otherwise
// If the user doesn't have a password, returns false
func (u *User) ComparePassword(password string) error {
	if len(u.Password) == 0 {
		return ErrUserHasNoPassword
	}
	if bcrypt.CompareHashAndPassword(u.Password, []byte(password)) != nil {
		return ErrIncorrectPassword
	}
	return nil
}

// GetIndexParams gets the indexParams.hjson file from the server for this user
// If it doesn't exist, will return default struct
func (u *User) GetIndexParams() (*params.MasterParams, error) {
	ip, err := database.GetIndexParams(u.Username)
	if err != nil {
		return params.DefaultMasterParams()
	}
	return ip, nil
}

// GetEcoParams gets the ecoParams.hjson file from the server for this user
func (u *User) GetEcoParams() (*params.EcoParams, error) {
	return database.GetEcoParams(u.Username)
}

// SaveMasterParams will overwrite the current index params file for this user with the given params
// Wraps database method
func (u *User) SaveMasterParams(ip *params.MasterParams) error {
	return database.SetMasterParams(u.Username, ip)
}

// SaveEcoParams will overwrite the current eco params file for this user with the given params
// Wraps database method
func (u *User) SaveEcoParams(ep *params.EcoParams) error {
	return database.SetEcoParams(u.Username, ep)
}

// ListJobs returns a list of the jobs the user has stored in the database
// Forwards the database function
func (u *User) ListJobs() []string {
	return database.ListJobs(u.Username)
}

// GetJob returns a job object that has the basic output and
// the status of the given job. If something unknown is queried will
// return a job with a failed status
func (u *User) GetJob(name string) (*Job, error) {
	j := &Job{Status: Failed}

	ip, ep, err := u.GetJobParams(name)
	if err != nil {
		return nil, fmt.Errorf("getting parameter files: %s", err)
	}

	// We don't mind if this can't be parsed - as we expect a failed job not to have this
	// file, or for it to be empty
	data, err := os.ReadFile(database.GetJobFilename(u.Username, name, FileJobOutput))
	if err == nil {
		j, err = parseJob(bytes.NewBuffer(data))
		if err != nil {
			return nil, fmt.Errorf("parsing job output: %w", err)
		}
	} else {
		if _, err = os.Stat(PathToJobFile(u.Username, name, FileJobProcessingFlag)); err == nil {
			j.Status = Processing
		}
	}

	j.Comment = ip.Comment
	j.TargetDatabase = ip.TargetDatabase
	j.Endpoint = ep.SaleEndpoint
	j.Name = name
	j.user = u
	return j, nil
}

// GetAllJobs gets every job this user has
func (u *User) GetAllJobs() ([]*Job, error) {
	jobIDs := u.ListJobs()
	jobs := make([]*Job, len(jobIDs))
	for idx, ID := range jobIDs {
		job, err := u.GetJob(ID)
		if err != nil {
			return nil, fmt.Errorf("getting job '%s': %w", ID, err)
		}
		jobs[idx] = job
	}
	return jobs, nil
}

// CreateJob creates a job out of the current context, saves it and returns the job
func (u *User) CreateJob(name string, ip *params.MasterParams, ep *params.EcoParams) (*Job, error) {
	var j = &Job{Name: name, user: u}

	if err := j.saveParams(ip, ep); err != nil {
		return nil, err
	}
	return j, nil
}

// DeleteJob will permantly remove a job
func (u *User) DeleteJob(name string) error {
	return database.DeleteJob(u.Username, name)
}

// GetJobParams will return the parameters used in the given job
// Will return error if the job doesn't exist or parameter files can't be loaded
func (u *User) GetJobParams(name string) (*params.MasterParams, *params.EcoParams, error) {
	mp, err := params.MasterParamsFromFile(database.GetJobFilename(u.Username, name, FileMasterFilename))
	if err != nil {
		return nil, nil, fmt.Errorf("reading master params: %w", err)
	}
	ep, err := params.EcoParamsFromFile(database.GetJobFilename(u.Username, name, FileEcoFilename))
	if err != nil {
		return nil, nil, fmt.Errorf("reading eco params: %w", err)
	}
	return mp, ep, nil
}

// SetJobParamsAsActive will set this jobs indexparams and ecoparams as the active index/eco params
func (u *User) SetJobParamsAsActive(name string) error {
	ip, ep, err := u.GetJobParams(name)
	if err != nil {
		return err
	}

	if err = database.SetMasterParams(u.Username, ip); err != nil {
		return fmt.Errorf("writing master params: %w", err)
	}
	if err = database.SetEcoParams(u.Username, ep); err != nil {
		return fmt.Errorf("writing eco params: %w", err)
	}
	return nil
}

// String returns the representation of a user for debugging
func (u *User) String() string {
	return fmt.Sprintf("%-16s%s\n%-16s%s %s\n%-16s%s\n%-16s%t", "Username:", u.Username, "Name:", u.Firstname, u.Surname, "Email:", u.Email, "Has Password:", u.Password != nil)
}

// Bytes returns JSON marshalled bytes
func (u *User) Bytes() ([]byte, error) {
	return json.Marshal(u)
}
