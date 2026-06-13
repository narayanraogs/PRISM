package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"prismServer/database"
	_ "prismServer/executeTest/measurements"
	"prismServer/logger"
	"prismServer/resultsDB"
	"prismServer/server"
	"prismServer/tc"
	"prismServer/utils"

	"github.com/gin-gonic/gin"
)

var cfgPath = flag.String("cfg", "~/prism/config/config.json", "Config File Path")
var portNo = flag.Int("port", 8080, "Port Number")

func init() {
	flag.Parse()
}

//go:embed web
var embeddedFiles embed.FS

var VersionString string = "Development"

func main() {
	fmt.Printf("Starting PRISM Server. Version: %s\n", VersionString)

	ok := utils.ReadConfiguration(*cfgPath)
	if !ok {
		return
	}
	tc.Init()
	utils.ReadSelectionParams()
	logPath := filepath.Join(utils.Config.BaseFolder, "log")
	logger.InitializeLog(logPath)
	createFolders()
	connectToDatabases()
	router := server.GetRouter()

	webFS, err := fs.Sub(embeddedFiles, "web")
	if err != nil {
		logger.Log.Error("Failed to access embedded web files", "error", err)
	}
	router.NoRoute(func(c *gin.Context) {
		http.FileServer(http.FS(webFS)).ServeHTTP(c.Writer, c.Request)
	})

	logger.Log.Info("Server started", "port", *portNo)
	err = router.Run(fmt.Sprintf(":%d", *portNo))
	if err != nil {
		logger.Log.Error("Server encountered an error while running", "error", err)
	}
}

func connectToDatabases() {
	ok := database.Connect(utils.Config.Database.DBPath)
	if !ok {
		logger.Log.Error("Cannot Connect to Database", "path", utils.Config.Database.DBPath)
		os.Exit(0)
	}
	resultsDB.Connect(utils.Config.Database.ResultsDBPath)
	tp, ok := database.GetSelectedTestPhase()
	if !ok {
		logger.Log.Warn("Unable to get selected Test Phase, defaulting to Unknown")
		tp = "Unknown"
	}
	utils.SetTestPhase(tp)
	utils.SetSatelliteName(utils.Config.SatelliteName)
}

func createFolders() {
	directoryName := utils.Config.BaseFolder + "/.resources"
	err := os.MkdirAll(directoryName, 0755)
	if err != nil {
		logger.Log.Error("Error creating directory", "error", err, "path", directoryName)
		return
	}
}
