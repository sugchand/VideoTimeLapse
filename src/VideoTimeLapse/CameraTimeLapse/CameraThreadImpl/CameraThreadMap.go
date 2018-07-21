package CameraThreadImpl

import (
    "sync"
    "VideoTimeLapse/dataSet"
    "VideoTimeLapse/CameraTimeLapse"
    "VideoTimeLapse/CameraTimeLapse/CameraThreadImpl/RTSPCameraImpl"
    "VideoTimeLapse/logging"
    "VideoTimeLapse/appErrors"
)

//The camera thread map is used by camerathreadrunner to keep track of running
// timelapse threads in the system.
type CameraThreadMap struct {
    threadMap map[string]CameraTimeLapse.CameraThreadInterface
    //Mutex to protect operation on above map.
    threadMapMutex sync.Mutex
}

//Adding camera thread to the map. Caller must make sure the entry is not
// present before adding to the map.
// ** NEVER INVOKE THIS FUNCTION FROM OUT OF THIS FILE **
func(t *CameraThreadMap)__addCameraThreadInMap(camName string,
                            threadObj CameraTimeLapse.CameraThreadInterface) {
    t.threadMapMutex.Lock()
    t.threadMap[camName] = threadObj
    t.threadMapMutex.Unlock()
}

// deleting the camera thread from the map.
// ** NEVER INVOKE THIS FUNCTION FROM OUT OF THIS FILE **
func(t *CameraThreadMap)__delCameraThreadInMap(camName string) {
    t.threadMapMutex.Lock()
    delete(t.threadMap, camName)
    t.threadMapMutex.Unlock()
}

//Return true and camera thread Obj if the camera thread obj present in the
// map.
// ** NEVER INVOKE THIS FUNCTION FROM OUT OF THIS FILE **
func(t *CameraThreadMap)__isCameraThreadInMap(camName string)(
                                  CameraTimeLapse.CameraThreadInterface, bool){
    t.threadMapMutex.Lock()
    obj, ret := t.threadMap[camName]
    t.threadMapMutex.Unlock()
    return obj,ret
}

//Concrete implementatation of thread generator interface.
// The generator just create a thread Obj. caller must initilize with
// proper values.
func(t *CameraThreadMap)PutCameraThreadObjInMap(cam *dataSet.Camera)(
                                 CameraTimeLapse.CameraThreadInterface, error) {
    log := logging.GetLoggerInstance()
    if(cam == nil || len(cam.Name) == 0) {
        log.Error("Invalid configuration, Cannot add camera to the map")
        return nil,appErrors.INVALID_INPUT
    }
    if obj,ret := t.__isCameraThreadInMap(cam.Name); ret == true {
        //Entry already in the map, no need to create new entry.
        return obj,appErrors.DATA_PRESENT_IN_SYSTEM
    }
    if cam.Type == dataSet.CAMERA_TYPE_RTSP {
        log.Trace("Creating RTSP camera thread for %s", cam.Name)
        var threadObj *RTSPCameraImpl.RTSPCameraThread
        threadObj = new(RTSPCameraImpl.RTSPCameraThread)
        t.__addCameraThreadInMap(cam.Name, threadObj)
        return threadObj, nil
    }
    return nil,appErrors.INVALID_INPUT
}

func(t *CameraThreadMap)RemoveCameraThreadObjInMap(camName string) {
    t.__delCameraThreadInMap(camName)
}

var camMapOnce sync.Once
var threadMapObj CameraThreadMap

func GetCameraMapObj() *CameraThreadMap {
    camMapOnce.Do(func() {
    //Init the dictionary of camera threads.
    threadMapObj.threadMap = make(map[string]CameraTimeLapse.CameraThreadInterface)
    })
    return &threadMapObj
}
