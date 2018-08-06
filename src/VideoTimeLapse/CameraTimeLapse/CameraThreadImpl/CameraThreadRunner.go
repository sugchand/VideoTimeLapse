package CameraThreadImpl

import (
    "sync"
    "VideoTimeLapse/logging"
    "VideoTimeLapse/appErrors"
    "VideoTimeLapse/config"
    "VideoTimeLapse/dataSet"
    "VideoTimeLapse/dataSet/dataSetImpl"
    "VideoTimeLapse/sys"
    "VideoTimeLapse/CameraTimeLapse"
)


type CameraThreadRunner struct {
    //Channel to receive camera update messages from other thread.
    camMsg chan *dataSet.Camera
}

const (
    MAX_CAMERA_MSG_SIZE = 50000
)

//function to validate if incoming message with camera status update.
func(camRunner *CameraThreadRunner)getCameraThreadOpSignal() *dataSet.Camera {
    select {
        case camObj := <-camRunner.camMsg:
            return camObj
        default:
            return nil
    }
}

// Function to do op on camera threads.
func(camRunner *CameraThreadRunner)SignalCameraThreadOp(
                            camera *dataSet.Camera) {
    camRunner.camMsg <- camera
}

//Function to iterate over all the camera threads and run the exit function.
func(camRunner *CameraThreadRunner)destroyAllCameraThreads() {
    var err error
    camMap := GetCameraMapObj()
    log := logging.GetLoggerInstance()
    for k,v := range camMap.threadMap {
        log.Trace("Destroying the camera thread %s", k)
        err = v.StopCameraThread()
        if err != nil {
            //Failed to stop the camera thread.
            log.Error("Failed to Stop the camera thread %s", k)
        } else {
            //Clear the threadObj from the threadMap if a thread killed
            //successfully.
            log.Trace("Successfully destroyed the camera thread %s", k)
            camMap.RemoveCameraThreadObjInMap(k)
        }
    }
}

// Function to update camera thread operation.
func(camRunner *CameraThreadRunner)executeCameraThread(
                              camera *dataSet.Camera, conf *config.AppConfig) {
    camMap := GetCameraMapObj()
    log := logging.GetLoggerInstance()
    camThread, err := camMap.PutCameraThreadObjInMap(camera)
    if err != nil {
        log.Trace("Failed to Create/Get camera thread for %s err : %s",
                                camera.Name, err)
        if err == appErrors.DATA_PRESENT_IN_SYSTEM {
            //The camera thread already present in system, hence
            //just update the threadas requested.
            log.Trace("The camera thread for %s already present in system",
                        camera.Name)
            //On update if camera got deleted/turned off, the thread has been
            //stopped and need to update the threadmap as well.
            if camera.Status != dataSet.CAMERA_STREAMING {
                err = camThread.StopCameraThread()
                if err != nil {
                    log.Error("Failed to kill camera thread %s, err %s",
                               camera.Name, err)
                    return
                }
                if camera.Status == dataSet.CAMERA_DELETED {
                    camMap.RemoveCameraThreadObjInMap(camera.Name)
                }
                return
            }
            //Camera thread specific fields are updated & is streaming.
            //camera parameter update are taken into account only when camera
            // camera is streaming.
            camThread.UpdateCameraThread(camera)
        }
        return
    }

    err = camThread.InitCameraThread(camera, conf)
    if err != nil {
        log.Error("Failed to initialize the camera thread obj %s, err:%s",
                    camera.Name, err)
        camMap.RemoveCameraThreadObjInMap(camera.Name)
        return
    }
    // Need to start a camera thread if it has status streaming.
    if camera.Status != dataSet.CAMERA_STREAMING {
        log.Trace("Streaming is not started on camera thread %s", camera.Name)
        return
    }
    // Camera thread is not present in the system, hence start the thread.
    camThread.RunCameraThread()
}


//Must be executed in a go routine.
func(camRunner *CameraThreadRunner)camThreadRunnerExecute(
                                                    conf *config.AppConfig) {
    syncObj := sys.GetAppSyncObj()
    log := logging.GetLoggerInstance()
    log.Trace("Starting cameraThreadRunner.")
    defer syncObj.ExitRoutineInWaitGroup()
    for {
        if syncObj.IsCameraThreadRunnerExited() {
            camRunner.destroyAllCameraThreads()
            break
        }
        camera := camRunner.getCameraThreadOpSignal()
        if  camera == nil {
            continue
        }
        camRunner.executeCameraThread(camera, conf)
    }
}

//Call the function only at the app startup. This function will start threads
// for all the camera configured in the system.
func(camRunner *CameraThreadRunner)CameraThreadRunnerStartup(
                                                conf *config.AppConfig) error {
    log := logging.GetLoggerInstance()
    dataObj := dataSetImpl.GetDataSetObj()
    cameras, err := dataObj.GetAllCameras()
    if err != nil {
        log.Error(`Cannot start camera threads , err:%s in
                    getting cameras from DB`, err)
        return err
    }
    camMap := GetCameraMapObj()
    var camThread CameraTimeLapse.CameraThreadInterface
    for _,camera := range cameras {
        if camera.Status == dataSet.CAMERA_STREAMING {
            //Start camera thread for all streaming cameras.
            camThread, err = camMap.PutCameraThreadObjInMap(&camera)
            if err != nil {
                log.Error(`Failed to get/create camera thread for %s
                             err : %s`, camera.Name, err)
                camMap.RemoveCameraThreadObjInMap(camera.Name)
                continue
            }
            //Initialize the camera thread first.
            err = camThread.InitCameraThread(&camera, conf)
            if err != nil {
                log.Error("Failed to initialize the camera thread obj %s, err:%s",
                    camera.Name, err)
            return err
            }
            //Start the camera thread
            camThread.RunCameraThread()
        }
    }
    return nil
}

func(camRunner *CameraThreadRunner)CamThreadRunnerMain(conf *config.AppConfig) {
    syncObj := sys.GetAppSyncObj()
    syncObj.AddRoutineInWaitGroup()
    //New runner routine to keep track of camera timelapse threads.
    go camRunner.camThreadRunnerExecute(conf)
}

var camRunnerOnce sync.Once
var camRunner CameraThreadRunner

func GetCameraThreadRunner() *CameraThreadRunner {
    camRunnerOnce.Do(func() {
        camRunner.camMsg = make(chan *dataSet.Camera, MAX_CAMERA_MSG_SIZE)
    })
    return &camRunner
}


