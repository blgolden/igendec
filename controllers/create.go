package controllers

import (
	"fmt"
	"strings"

	"github.com/blgolden/igendec/epds"
	"github.com/blgolden/igendec/params"

	"github.com/blgolden/igendec/logger"

	"github.com/gofiber/fiber/v2"
)

// Create will render a menu which allows you to either select an old job to run
// or build a new job
func (h *Handler) Create(c *fiber.Ctx) error {
	user, err := h.Session.User(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}

	m := make(fiber.Map)
	jobs, err := user.GetAllJobs()
	if err != nil {
		return fmt.Errorf("getting all jobs: %w", err)
	}
	m["Jobs"] = jobs
	m["Endpoints"] = params.EndpointSlice
	m["IndexTypes"] = params.IndexTypes
	m["Databases"] = epds.ListDatabases()

	return h.RenderPrimary("create", m, c)
}

// CreateBuild is the endpoint for getting create/build page
// create/build page allows you to set parameters for iGenDec run
// It has two query parameters:
// endpoint: the endpoint to build from. This must be set if job is not set
// job: the job to use for getting the query params
// If job is set, it will ignore anything set in the endpoint query parameter
func (h *Handler) CreateBuild(c *fiber.Ctx) error {
	user, err := h.Session.User(c)
	if err != nil {
		return ErrInternalServer
	}

	var (
		masterParams *params.MasterParams
		ecoParams    *params.EcoParams
	)

	endpoint, endpointOK := params.EndpointMap[c.Query("endpoint")]
	indextype := params.IndexType(c.Query("indextype"))
	jobname := c.Query("job")
	if (!endpointOK || indextype == "") && jobname == "" {
		return fiber.NewError(fiber.StatusBadRequest, "not all expected parameters found")
	}

	// If jobname was given in the query load that jobs parameters
	if jobname != "" {
		masterParams, ecoParams, err = user.GetJobParams(jobname)
		if err != nil {
			return ErrInternalServer
		}

		// Otherwise load the current index/ecoparams
	} else {
		masterParams, err = params.DefaultMasterParams()
		if err != nil {
			return ErrInternalServer
		}
		ecoParams, err = params.DefaultEcoParams(endpoint, indextype)
		if err != nil {
			return ErrInternalServer
		}
		ecoParams.IndexTerminal = indextype == params.Terminal
	}

	// If a target database has been given, set the IndexComponents defaults to be the available keys
	// in the selected database
	if db, err := epds.NewDatabase(c.Query("target-database")); err == nil {
		ecoParams.IndexComponents = db.TraitKeys(ecoParams.IndexComponents)
		masterParams.TargetDatabase = db.Name
	}

	// Set these to the users values
	if err = user.SaveEcoParams(ecoParams); err != nil {
		return fmt.Errorf("saving eco params: %w", err)
	}
	// Set these to the users values
	if err = user.SaveMasterParams(masterParams); err != nil {
		return fmt.Errorf("saving master params: %w", err)
	}

	// Load in the params to a map we use for rendering html
	m := make(fiber.Map)
	m = masterParams.ToMap(m)
	m = ecoParams.ToMap(m)

	// Need to change PlanningHorizon if the indextype is terminal
	if indextype == params.Terminal {
		if endpoint == params.Weaning {
			m["PlanningHorizon"] = 1
		} else {
			m["PlanningHorizon"] = 2
		}
	}

	return h.RenderPrimary("create-build", m, c)
}

// CreateUpdate updates a users parameters files via POST
// The form key should match a value
func (h *Handler) CreateUpdate(c *fiber.Ctx) error {
	user, err := h.Session.User(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}

	// Get parameters for this user
	ip, err := user.GetIndexParams()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}
	ep, err := user.GetEcoParams()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}

	// Try parse into structs
	if err = c.BodyParser(ip); err != nil {
		fmt.Println(err)
		return c.Status(fiber.StatusBadRequest).SendString("Could not read values")
	}
	if err = c.BodyParser(ep); err != nil {
		fmt.Println(err)
		return c.Status(fiber.StatusBadRequest).SendString("Could not read values")
	}

	// Save the parameters back to the server
	if err = user.SaveMasterParams(ip); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}
	if err = user.SaveEcoParams(ep); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}

	return c.SendStatus(fiber.StatusOK)
}

// CreateSubmit runs a job through iGenDecModel
func (h *Handler) CreateSubmit(c *fiber.Ctx) error {
	user, err := h.Session.User(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}

	ip, err := user.GetIndexParams()
	if err != nil {
		return ErrInternalServer
	}
	ep, err := user.GetEcoParams()
	if err != nil {
		return ErrInternalServer
	}

	// Save the comment
	ip.Comment = c.FormValue("comment")

	jobname := strings.TrimSpace(c.FormValue("name"))
	if !NameRegex.MatchString(jobname) {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid job name, can only contain letters, numbers, and special characters '-', '_'")
	}

	job, err := user.CreateJob(jobname, ip, ep)
	if err != nil {
		logger.Debug("%s", err)
		return ErrInternalServer
	}
	if err = job.Run(); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to run job. Please contact support")
	}
	return c.SendStatus(fiber.StatusOK)
}

// CreateRun will run a given job
func (h *Handler) CreateRun(c *fiber.Ctx) error {
	user, err := h.Session.User(c)
	if err != nil {
		logger.Debug("%s", err)
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}

	jobname := c.FormValue("job")

	job, err := user.GetJob(jobname)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}

	if err = job.Run(); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to run job. Please contact support")
	}
	return c.SendStatus(fiber.StatusOK)
}
