package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func GetRouter() *gin.Engine {
	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/serverStatus", serverStatus)
	//RFUplinkRelated
	r.POST("/getRFUplinkMetaData", getRFUplinkMetaData)
	r.POST("/getLinkStatus", getLinkStatus)
	r.POST("/setTSMRoute", setTSMRoute)
	r.POST("/setTSMAttn", setTSMAttn)
	//TestRelated
	r.POST("/getAllTests", getAllTests)
	r.GET("/testProgress", testProgress)
	//StabilityRelated
	r.POST("/getStabilityMetadata", getStatbilityMetadata)
	//SpectrumDumpRelated
	r.POST("/getSpectrumDumpMetadata", getSpectrumDumpMetadata)
	r.POST("/setSpectrum", setSpectrum)
	r.POST("/readSpectrum", readSpectrum)
	r.POST("/dumpSpectrum", dumpSpectrun)
	r.POST("/saveSpectrum", saveSpectrum)
	r.POST("/dumpTrace", dumpTrace)
	r.POST("/dumpScreenshot", dumpScreenshot)
	//MonitorRelated
	r.POST("/getMonitorMetadata", getMonitorMetadata)
	r.GET("/monitor", monitor)
	//TVACCableLossRelated
	r.POST("/getTVACCableLossMetadata", getTVACCableLossMetadata)
	r.POST("/getTVACCableMeasuredDetails", getTVACCableMeasuredDetails)
	r.GET("/measureTVACCableLoss", measureTVACCableLoss)
	return r
}
