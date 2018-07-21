package CameraTimeLapse

import (
    "VideoTimeLapse/dataSet"
    "VideoTimeLapse/config"
)

//Define interface for camera threads. Different type of cameras will have
// their own implementations.
type CameraThreadInterface interface {
    InitCameraThread(*dataSet.Camera, *config.AppConfig) (error)
    RunCameraThread() error
    UpdateCameraThread(*dataSet.Camera)(error)
    StopCameraThread() error
}
