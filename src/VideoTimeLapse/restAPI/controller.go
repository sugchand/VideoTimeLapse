// Copyright 2018 Sugesh Chandran
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package restAPI

import (
    "net/http"
    "encoding/json"
    "io"
    "io/ioutil"
    "github.com/gorilla/mux"
    "VideoTimeLapse/dataSet"
    "VideoTimeLapse/dataSet/dataSetImpl"
    "VideoTimeLapse/CameraTimeLapse/CameraThreadImpl"
    "VideoTimeLapse/logging"
    "VideoTimeLapse/appErrors"
)

type controller struct { }

func (ctrl *controller) getAllCameras(w http.ResponseWriter, r *http.Request) {
    log := logging.GetLoggerInstance()
    dataObj := dataSetImpl.GetDataSetObj()
    if dataObj == nil {
        log.Error("Empty datastore handle")
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("500-Server Error "))
    }
    rows, err := dataObj.GetAllCameras()
    if err != nil {
        log.Trace("Failed to get the records from DB")
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("500-Server Error "+ err.Error()))
        return
    }
    data, _ := json.Marshal(rows)
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.WriteHeader(http.StatusOK)
    w.Write(data)
}

func(ctrl *controller)signalCameraThreadRunner(camera *dataSet.Camera) error {
    var err error
    if err == nil {
        //Successfully added a camera to the system.
        // Must signal the camerathread runner about the new camera.
        camRunner := CameraThreadImpl.GetCameraThreadRunner()
        //It is very unlikely the signal on a channel can cause
        // the thread to block, as the camerathreadrunner can hold
        //up the messages and let the thread continue.
        camRunner.SignalCameraThreadOp(camera)
    }
    return err
}

func (ctrl *controller) createCamera(w http.ResponseWriter, r *http.Request) {
    log := logging.GetLoggerInstance()
    body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
    if err != nil {
        log.Error("Failed to read request,")
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    if err := r.Body.Close(); err != nil {
        log.Error("Failed to close the request.")
    }
    jsonCamObj := new(JsonCameraInput)
    jsonCamObj.AllocateFields()
    if err := json.Unmarshal(body, &jsonCamObj); err != nil {
        w.WriteHeader(422)
        log.Error("Failed to Unmarshal the camera input err:%s", err)
        if err := json.NewEncoder(w).Encode(err); err != nil {
            log.Error("Failed to encode marshaling err : %s", err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
    }

    dataObj := dataSetImpl.GetDataSetObj()
    var camObj dataSet.Camera
    jsonCamObj.ReadJsonData(&camObj) //Read json data to original camera obj
    err = dataObj.AddNewCamera(&camObj)
    if err != nil {
        log.Error("Failed to create camera entry in table err :%s", err)
        if err == appErrors.DATA_PRESENT_IN_SYSTEM {
            w.WriteHeader(400) //Bad Request.
            return
        }
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusCreated)

    //Update the camerathreadRunner with new camera created
    //The request may have partial information on the camera Obj.
    //So update the camerathread runner using the entry stored in the system.
    var camInSys *dataSet.Camera
    camInSys, err = dataObj.GetCamera(camObj.Name)
    if err != nil || camInSys == nil {
        log.Error(`camerathreadrunner update failed as newely created camera
                   entry %s not found in system`, camObj.Name)
        return
    }
    ctrl.signalCameraThreadRunner(camInSys)
}

func (ctrl *controller) getCamera(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    log := logging.GetLoggerInstance()
    cameraId := vars["camera-name"]
    dataObj := dataSetImpl.GetDataSetObj()
    if len(cameraId) == 0 {
        log.Error("Empty camera ID , cannot find it")
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    camObj, err := dataObj.GetCamera(cameraId)
    if err != nil {
        log.Error("Failed to get Camera from the server err:%s", err)
        if err == appErrors.DATA_NOT_FOUND {
            w.WriteHeader(http.StatusNotFound)
            return
        }
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    data, _ := json.Marshal(camObj)
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.WriteHeader(http.StatusOK)
    w.Write(data)
}

func (ctrl *controller) deleteCamera(w http.ResponseWriter, r *http.Request) {
    var err error
    vars := mux.Vars(r)
    log := logging.GetLoggerInstance()
    cameraId := vars["camera-name"]
    dataObj := dataSetImpl.GetDataSetObj()
    if len(cameraId) == 0 {
        log.Error("Empty camera ID , cannot find it")
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    var camObj *dataSet.Camera
    camObj,err = dataObj.GetCamera(cameraId)
    if err != nil || camObj == nil {
        log.Error(`Failed to retrieive the camera object, cannot delete %s
                    err : %s`, cameraId, err)
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    err = dataObj.DeleteCamera(cameraId)
    if err != nil {
        log.Error("Failed to delete the camera err : %s", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
    //Destroy the camera thread as its deleted.
    camObj.Status = dataSet.CAMERA_DELETED
    ctrl.signalCameraThreadRunner(camObj)
}

func (ctrl *controller) updateCamera(w http.ResponseWriter, r *http.Request) {
    var err error
    vars := mux.Vars(r)
    log := logging.GetLoggerInstance()
    cameraId := vars["camera-name"]
    if len(cameraId) == 0 {
        log.Error("Empty camera ID , cannot find it")
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
    if err != nil {
        log.Error("Failed to read request,")
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    if err := r.Body.Close(); err != nil {
        log.Error("Failed to close the request.")
    }

    jsonCamObj := new(JsonCameraInput)
    jsonCamObj.AllocateFields()
    if err := json.Unmarshal(body, &jsonCamObj); err != nil {
        w.WriteHeader(422)
        log.Error("Failed to Unmarshal the camera input err:%s", err)
        if err := json.NewEncoder(w).Encode(err); err != nil {
            log.Error("Failed to encode marshaling err : %s", err)
            w.WriteHeader(http.StatusBadRequest)
            return
        }
    }
    if len(*jsonCamObj.Name) == 0 {
        //Patch request doesnt need to populate all the fields in json.
        *jsonCamObj.Name = cameraId
    }
    if *jsonCamObj.Name != cameraId {
        //Wrong camera ID
        w.WriteHeader(http.StatusBadRequest)
    }

    dataObj := dataSetImpl.GetDataSetObj()
    var camObj *dataSet.Camera
    //Get the record first before trying to update.
    camObj, err = dataObj.GetCamera(cameraId)
    if err != nil {
        log.Error("Failed to get the record, Cannot update the camera %s, err:%s",
                    *jsonCamObj.Name, err)
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    jsonCamObj.ReadJsonData(camObj)
    err = dataObj.UpdateCamera(camObj)
    if err != nil {
        log.Error("Failed to update the camera %s", err)
        if err == appErrors.DATA_NOT_FOUND {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
}

func (ctrl *controller) getVideos(w http.ResponseWriter, r *http.Request) {
}

func (ctrl *controller) getVideo(w http.ResponseWriter, r *http.Request) {
}

func (ctrl *controller) deleteVideos(w http.ResponseWriter, r *http.Request) {
}

func (ctrl *controller) deleteVideo(w http.ResponseWriter, r *http.Request) {
}
