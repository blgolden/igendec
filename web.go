package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/blgolden/igendec/epds"

	"github.com/blgolden/igendec/params"

	"github.com/blgolden/igendec/controllers"
	"github.com/blgolden/igendec/logger"
	"github.com/blgolden/igendec/routes"
	"github.com/blgolden/igendec/users"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html"
)

// CLI stuff
var (
	version           = "0.0.3"
	port              = kingpin.Flag("port", "Port to listen on").Short('p').Default("3000").Int()
	addr              = kingpin.Flag("addr", "Address to listen on").Short('a').Default("localhost").String()
	defaultMasterPath = kingpin.Flag("master-path", "Path to the default master parameter file").Default("./defaultMaster.hjson").String()

	defaultWeaningPath     = kingpin.Flag("eco-weaning-path", "Path to the default economic index parameter file for weaning").Default("./defaultEcoWeaning.hjson").String()
	defaultWeaningTermPath = kingpin.Flag("eco-weaning-term-path", "Path to the default economic index parameter file for weaning for a terminal index").Default("./defaultEcoWeaningTerm.hjson").String()

	defaultBackgroundPath     = kingpin.Flag("eco-background-path", "Path to the default economic index parameter file for weaning").Default("./defaultEcoBackground.hjson").String()
	defaultBackgroundTermPath = kingpin.Flag("eco-background-term-path", "Path to the default economic index parameter file for weaning for a terminal index").Default("./defaultEcoBackgroundTerm.hjson").String()

	defaultFatPath     = kingpin.Flag("eco-fat-path", "Path to the default economic index parameter file for fat cattle").Default("./defaultEcoFatcattle.hjson").String()
	defaultFatTermPath = kingpin.Flag("eco-fat-term-path", "Path to the default economic index parameter file for fat cattle for a terminal index").Default("./defaultEcoFatcattleTerm.hjson").String()

	defaultSlaughterPath     = kingpin.Flag("eco-slaughter-path", "Path to the default economic index parameter file for slaughter cattle").Default("./defaultEcoSlaughtercattle.hjson").String()
	defaultSlaughterTermPath = kingpin.Flag("eco-slaughter-term-path", "Path to the default economic index parameter file for slaughter cattle for a terminal index").Default("./defaultEcoSlaughtercattleTerm.hjson").String()

	databaseDirectory = kingpin.Flag("bull-database", "Path to the directory containing all of the epds for running jobs against").Short('d').Default("./epds").String()
	userBlacklist     = kingpin.Flag("user-blacklist", "Path to file containing a list of names. These users will not be authenticated on login").Short('b').Default("./user-blacklist.txt").String()

	usersPath = kingpin.Flag("users-path", "Path to location where users' accounts are stored").Short('u').Default("/tmp/igendecDB").String()
)

// Initilises objects and environment
func setup(ctx context.Context) {
	// Init singleton packages
	logger.Init()

	users.UsersPath = *usersPath
	users.Init()

	h := controllers.NewHandler()

	// Set the default paths
	params.DefaultMasterPath = *defaultMasterPath

	params.DefaultWeaningPath = *defaultWeaningPath
	params.DefaultWeaningTerminalPath = *defaultWeaningTermPath

	params.DefaultBackgroundPath = *defaultBackgroundPath
	params.DefaultBackgroundTerminalPath = *defaultBackgroundTermPath

	params.DefaultFatPath = *defaultFatPath
	params.DefaultFatTerminalPath = *defaultFatTermPath

	params.DefaultSlaughterPath = *defaultSlaughterPath
	params.DefaultSlaughterTerminalPath = *defaultSlaughterTermPath

	epds.DatabasePath = *databaseDirectory

	// Check if there is a blacklist - if so load in
	if info, err := os.Stat(*userBlacklist); err == nil && !info.IsDir() {
		data, err := ioutil.ReadFile(*userBlacklist)
		if err != nil {
			logger.Fatal("trying to read in blacklist: %s", err)
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			h.UserBlacklist[strings.TrimSpace(line)] = struct{}{}
		}
	}

	// Set templating engine to html templates
	// Keeping templates in folder 'views'
	engine := html.New("./views", ".html")
	engine.AddFunc("json", func(i interface{}) string {
		data, _ := json.Marshal(i)
		return string(data)
	})

	// Create app
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// Set static files folder to folder 'public'
	app.Static("/", "./public")

	// Middleware
	app.Use(favicon.New(favicon.Config{
		File: "./public/favicon.ico",
	})) // handles favicons requests nicely
	app.Use(fiberlogger.New())
	app.Use(cors.New()) // For enabling CORS
	app.Use(h.Authorise)

	routes.RegisterRoutes(app, h)

	app.Use(h.NotFound)

	// Listen on port 3000
	go func() {
		app.Listen(fmt.Sprintf("%s:%d", *addr, *port))
	}()

	<-ctx.Done()

	// Handle closing down systems, backing up data
	fmt.Println("Handle closing down systems here")
}

// We create a context here to run the web server in
// This means if any OS signals are sent to the program, we can delay them and save any state and make sure everything closes down correctly
func main() {
	// Parse args
	kingpin.Version(version)
	kingpin.Parse()

	// Create channel to listen for os signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Set up context
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		oscall := <-c
		log.Printf("system call:%+v", oscall)
		cancel()
	}()

	// Start server
	setup(ctx)
}
