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
    "VideoTimeLapse/logging"
)

type controller struct {
    dataObj dataSet.DataSetInterface
}

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

    var camObj dataSet.Camera
    if err := json.Unmarshal(body, &camObj); err != nil {
        w.WriteHeader(422)
        log.Error("Failed to Unmarshal the camera input err:%s", err)
        if err := json.NewEncoder(w).Encode(err); err != nil {
            log.Error("Failed to encode marshaling err : %s", err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
    }

    dataObj := dataSetImpl.GetDataSetObj()
    err = dataObj.AddNewCamera(&camObj)
    if err != nil {
        log.Error("Failed to create camera entry in table err :%s", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusCreated)
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
}

func (ctrl *controller) updateCamera(w http.ResponseWriter, r *http.Request) {
}

func (ctrl *controller) getVideos(w http.ResponseWriter, r *http.Request) {
}

func (ctrl *controller) getVideo(w http.ResponseWriter, r *http.Request) {
}

func (ctrl *controller) deleteVideos(w http.ResponseWriter, r *http.Request) {
}

func (ctrl *controller) deleteVideo(w http.ResponseWriter, r *http.Request) {
}
