package controllers

import (
	"fmt"
	"strings"

	"github.com/blgolden/igendec/epds"

	"github.com/gofiber/fiber/v2"
)

// Jobs renders the jobs page
// Gets user, and returns the render function with the users Jobs passed to it
func (h *Handler) Jobs(c *fiber.Ctx) error {
	user, err := h.Session.User(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}
	jobs := user.ListJobs()
	var defaultJob string
	if len(jobs) > 0 {
		defaultJob = jobs[0]
	}

	return h.RenderPrimary("jobs", fiber.Map{"JobsList": jobs, "Selected": c.Query("job", defaultJob)}, c)
}

// JobsInfo returns the html for a job
func (h *Handler) JobsInfo(c *fiber.Ctx) error {
	user, err := h.Session.User(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}
	var name = c.Query("name")

	job, err := user.GetJob(name)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("bad job name")
	}

	err = c.Render("jobs/jobdetails", fiber.Map{"Job": job})
	return err
}

// JobsDownload will return the zip compressed files for a job
func (h *Handler) JobsDownload(c *fiber.Ctx) error {
	user, err := h.Session.User(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}
	var jobName = c.Query("id")

	if jobName == "" {
		return c.Status(fiber.StatusBadRequest).SendString("invalid job name")
	}

	job, err := user.GetJob(jobName)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("bad job name")
	}

	zippedData, err := job.Zip()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}

	c.Append(fiber.HeaderContentType, "application/zip")
	c.Append(fiber.HeaderContentDisposition, `attachment; filename="`+jobName+`.zip"`)
	return c.Send(zippedData)
}

// JobsDelete will delete the given job
func (h *Handler) JobsDelete(c *fiber.Ctx) error {
	user, err := h.Session.User(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}
	var jobName = c.Query("id")

	if jobName == "" || jobName == "." || jobName == ".." || strings.Contains(jobName, "/") {
		return c.Status(fiber.StatusBadRequest).SendString("invalid job name")
	}

	if err := user.DeleteJob(jobName); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}
	return nil
}

// JobsSelect renders the page showing the databases - allowing a user to compare
func (h *Handler) JobsSelect(c *fiber.Ctx) error {
	if c.Query("job") == "" {
		return h.NotFound(c)
	}

	user, err := h.Session.User(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}
	selected := c.Query("target-database")
	dbs := epds.ListDatabases(user.Access)
	found := false
	for _, dbName := range dbs {
		if dbName == selected {
			found = true
			break
		}
	}
	if !found && len(dbs) > 0 {
		selected = dbs[0]
	}
	return h.RenderPrimary("jobs-select", fiber.Map{"Databases": dbs, "Job": c.Query("job"), "Selected": selected}, c)
}

// JobsSelectDatabase will render the database information partial
// to be injected into the database select page
func (h *Handler) JobsSelectDatabase(c *fiber.Ctx) error {
	user, err := h.Session.User(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}

	database, err := epds.NewDatabase(c.Query("name"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "bad database name")
	}

	job, err := user.GetJob(c.Query("job"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "bad job name")
	}
	for _, ic := range job.Output {
		if v, ok := database.Xref[ic.Key()]; ok {
			v.Select = true
			database.Xref[ic.Key()] = v
		}
	}

	return c.Render("jobs/databasedetails", fiber.Map{"Database": database, "Fields": database.FieldSlice(), "JobName": job.Name})
}

// JobsSelectDatabaseCompare is called when the user wants to compare a job to a database
func (h *Handler) JobsSelectDatabaseCompare(c *fiber.Ctx) error {
	user, err := h.Session.User(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}

	job, err := user.GetJob(c.FormValue("job"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("bad job name")
	}

	database, err := epds.NewDatabase(c.FormValue("name"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("bad database name")
	}

	var values []string
	c.Context().PostArgs().VisitAll(func(key []byte, val []byte) {
		if string(key) == "job" || string(key) == "name" {
			return
		}
		values = append(values, string(key))
	})

	buf, err := database.CompareJob(job, values)
	if err != nil {
		fmt.Println(err)
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}

	c.Append(fiber.HeaderContentType, "application/text")
	c.Append(fiber.HeaderContentDisposition, `attachment; filename="compare.csv"`)

	return c.Send(buf.Bytes())
}
