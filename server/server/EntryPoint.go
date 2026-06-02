package server

import (
	"prismServer/global"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var gbl global.ClientGlobal

func GetRouter() *gin.Engine {
	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/bootstrapData", getBootstrapData)

	r.GET("/serverStatus", serverStatus)
	//RFUplinkRelated
	r.POST("/getLinkStatus", getLinkStatus)
	r.POST("/setTSMRoute", setTSMRoute)
	r.POST("/setTSMAttn", setTSMAttn)
	//TestRelated
	r.GET("/testProgress", testProgress)
	//StabilityRelated
	r.GET("/stability", stability)
	r.POST("/getStabilityPoints", getStabilityPoints)
	r.POST("/getStabilityReports", getStabilityReports)
	//SpectrumDumpRelated
	r.POST("/setSpectrum", setSpectrum)
	r.POST("/readSpectrum", readSpectrum)
	r.POST("/dumpSpectrum", dumpSpectrun)
	r.POST("/saveSpectrum", saveSpectrum)
	r.POST("/dumpTrace", dumpTrace)
	r.POST("/dumpScreenshot", dumpScreenshot)
	//MonitorRelated
	r.GET("/monitor", monitor)
	//TVACCableLossRelated
	r.POST("/getTVACCableMeasuredDetails", getTVACCableMeasuredDetails)
	r.GET("/measureTVACCableLoss", measureTVACCableLoss)
	//CableLossRelated
	r.POST("/getCableMeasuredDetails", getCableMeasuredDetails)
	r.GET("/measureCableLoss", measureCableLoss)
	//AttnRelated
	r.GET("/measureAttn", measureAttn)
	//DatabaseRelated
	r.POST("/getConfigsForUplink", getConfigsForUplink)
	r.POST("/getConfigsForDownlink", getConfigsForDownlink)
	r.POST("/getUplinkLossProfile", getUplinkLossProfile)
	r.POST("/getDownlinkLossProfile", getDownlinkLossProfile)
	r.POST("/saveUplinkLossProfile", saveUplinkLossProfile)
	r.POST("/saveDownlinkLossProfile", saveDownlinkLossProfile)
	r.POST("/selectTestPhase", selectTestPhase)
	r.POST("/addNewTestPhase", addNewTestPhase)
	//ResultRelated
	r.POST("/getResultMetadata", getResultMetadata)
	r.POST("/getReportPDF", getReportPDF)
	r.POST("/regenerateReport", regenerateReport)
	r.POST("/getReportsData", getReportsData)
	//TSMInternalLossRelated
	r.GET("/measureTSMInternalLoss", measureTSMInternalLoss)
	r.POST("/createNewTSMTable", createNewTSMTable)
	//GTxMeasurementRelated
	r.GET("/conductGTxTne", conductGTxTne)
	//UpDownConverterRelated
	r.GET("/upDownConverterMeasurement", upDownConverterMeasurement)
	r.GET("/getUCDCResults", getUCDCResults)
	r.POST("/upDownConverterResult", upDownConverterResult)
	//SCPI
	r.GET("/scpi", scpi)
	//PlannerState
	r.POST("/savePlannerData", savePlannerData)
	r.POST("/loadPlannerData", loadPlannerData)
	return r
}
