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
	r.GET("/stability", stability)
	r.POST("/getStabilityReportsMetadata", getStabilityReportsMetadata)
	r.POST("/getStabilityPoints", getStabilityPoints)
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
	//CableLossRelated
	r.POST("/getCableLossMetadata", getCableLossMetadata)
	r.POST("/getCableMeasuredDetails", getCableMeasuredDetails)
	r.GET("/measureCableLoss", measureCableLoss)
	//AttnRelated
	r.POST("/getAttnMetadata", getAttnMetadata)
	r.GET("/measureAttn", measureAttn)
	//DatabaseRelated
	r.POST("/getDatabaseMetadata", getDatabaseMetadata)
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
	//TSMInternalLossRelated
	r.POST("/getTSMInternalLossMetadata", getTSMInternalLossMetadata)
	r.GET("/measureTSMInternalLoss", measureTSMInternalLoss)
	return r
}
