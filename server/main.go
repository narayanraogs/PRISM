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
		fmt.Println(err)
	}
	router.NoRoute(func(c *gin.Context) {
		http.FileServer(http.FS(webFS)).ServeHTTP(c.Writer, c.Request)
	})

	fmt.Printf("Server started on PORT %d\n", *portNo)
	err = router.Run(fmt.Sprintf(":%d", *portNo))
	if err != nil {
		fmt.Println(err)
	}
}

func connectToDatabases() {
	ok := database.Connect(utils.Config.Database.DBPath)
	if !ok {
		fmt.Println("Cannot Connect to Database")
		os.Exit(0)
	}
	resultsDB.Connect(utils.Config.Database.ResultsDBPath)
	tp, ok := database.GetSelectedTestPhase()
	if !ok {
		logger.Log.Info("Unable to get selected Test Phase")
		tp = "Unknown"
	}
	utils.SetTestPhase(tp)
	utils.SetSatelliteName(utils.Config.SatelliteName)
}

func createFolders() {
	directoryName := utils.Config.BaseFolder + "/.resources"
	err := os.MkdirAll(directoryName, 0755)
	if err != nil {
		fmt.Println("Error creating directory", err)
		return
	}
}
