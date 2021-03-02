package users

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/hjson/hjson-go"
	"github.com/klauspost/compress/zip"
	"github.com/blgolden/igendec/params"
)

type trait string

// Types of traits
const (
	TraitWW   = trait("WW")
	TraitMW   = trait("MW")
	TraitSTAY = trait("STAY")
	TraitCD   = trait("CD")
	TraitHP   = trait("HP")
)

func (t trait) String() string {
	return string(t)
}

type component string

// Types of components
const (
	ComponentDirect   = component("D")
	ComponentMaternal = component("M")
)

func (c component) String() string {
	switch c {
	case ComponentDirect:
		return "Direct"
	case ComponentMaternal:
		return "Maternal"
	}
	return "Unknown"
}

// JobStatus are the possible states a job can have
type JobStatus string

// Statuses for a job
var (
	Passed     = JobStatus("passed")
	Failed     = JobStatus("failed")
	Processing = JobStatus("processing")
)

// Job holds information on a job run through create page
type Job struct {
	user           *User
	Name           string
	Status         JobStatus
	Endpoint       string
	Output         []IndexElement `json:"indexElement"`
	Comment        string
	TargetDatabase string
}

// IndexElement holds the details of a trait thats output from iGenDec
type IndexElement struct {
	Trait                 trait     `json:"trait"`
	Component             component `json:"component"`
	MarginalEconomicValue float64   `json:"mev"`
	DisplayMEV            string
}

// Key returns the trait and component joined by a comma to conform to how we otherwise find this
func (ic IndexElement) Key() string {
	return string(ic.Trait) + "," + string(ic.Component)
}

// ParseJob parses a job from a reader
func parseJob(r io.Reader) (*Job, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	elements := make(map[string]interface{})
	if err = hjson.Unmarshal(data, &elements); err != nil {
		return nil, err
	}
	job := &Job{}
	if data, err = json.Marshal(elements); err != nil {
		return nil, err
	}
	if err = json.Unmarshal(data, job); err != nil {
		return nil, err
	}
	job.Status = Passed

	for idx := range job.Output {
		job.Output[idx].DisplayMEV = strconv.FormatFloat(job.Output[idx].MarginalEconomicValue, 'f', 3, 64)
	}

	return job, nil
}

// Run uses the exec package to run the iGenDec job with the starters binary
// starters needs to be in the path
func (job *Job) Run() error {
	os.Create(PathToJobFile(job.user.Username, job.Name, FileJobProcessingFlag))
	cmd := exec.Command("starter", "-genParm", PathToJobFile(job.user.Username, job.Name, FileMasterFilename), "-indexParm", PathToJobFile(job.user.Username, job.Name, FileEcoFilename), "-outputFile", PathToJobFile(job.user.Username, job.Name, FileJobOutput))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running job: %w", err)
	}
	os.Remove(PathToJobFile(job.user.Username, job.Name, FileJobProcessingFlag))
	return nil
}

// Zip compresses all the job files and returns the zipped archive as bytes
func (job *Job) Zip() ([]byte, error) {
	buf := &bytes.Buffer{}
	w := zip.NewWriter(buf)

	err := filepath.Walk(PathToJobFile(job.user.Username, job.Name, ""), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		f, err := w.Create(info.Name())
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)
		return err
	})

	if err != nil {
		return nil, err
	}
	if err = w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// saves the parameter files for this job
func (job *Job) saveParams(ip *params.MasterParams, ep *params.EcoParams) error {
	data, err := ip.Bytes()
	if err != nil {
		return fmt.Errorf("encoding master params: %w", err)
	}
	if err := os.WriteFile(database.GetJobFilename(job.user.Username, job.Name, FileMasterFilename), data, 0755); err != nil {
		return fmt.Errorf("writing master params: %w", err)
	}

	data, err = ep.Bytes()
	if err != nil {
		return fmt.Errorf("encoding eco params: %w", err)
	}
	if err := os.WriteFile(database.GetJobFilename(job.user.Username, job.Name, FileEcoFilename), data, 0755); err != nil {
		return fmt.Errorf("writing eco params: %w", err)
	}
	return nil
}
