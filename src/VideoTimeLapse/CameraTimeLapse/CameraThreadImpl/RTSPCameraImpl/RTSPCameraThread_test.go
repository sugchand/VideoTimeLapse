package RTSPCameraImpl

// Test file for validating the RTSP camera handler functions.
import (
    "testing"
    "VideoTimeLapse/dataSet"
    "VideoTimeLapse/config"
    "VideoTimeLapse/logging"
)

func TestCompactTimeLapseVideo(t *testing.T) {
    fileName := "./test.mp4"
    logger := new(logging.Logging)
    logger.LogInitSingleton(logging.LogLeveltype(logging.Trace),
                            "")
    logger.Trace("Logging service is started..")
    var camThread RTSPCameraThread
    var camThreadptr *RTSPCameraThread
    camThreadptr = &camThread
    var cam dataSet.Camera
    var conf config.AppConfig
    camThreadptr.InitCameraThread(&cam, &conf)
    camThreadptr.compactTimeLapseVideo(fileName)
    t.Log("Completed the videoTimeLapseCompact ")
}
