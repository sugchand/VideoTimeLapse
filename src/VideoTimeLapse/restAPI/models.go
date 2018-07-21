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
    "VideoTimeLapse/dataSet"
)

// Must be updated when original camera structure is modified. This structure
// has one to one mapping to original camera structure. The camera parameters
// that are exposed to outside world for input are only included in this Json.
// When deseralizing the data from json, application must know if the field is
// being notprovided or empty value. Json unmarshalling set all the unprovided
// values to its default. i.e for string it is "", for integer its 0 and so on.
// So to distinguish default from unprovided values, a model struct is being
// used instead of using original structure for unmarshalling. The new model
// keep 1:1 mapping to all the fields in original structure except their
// datatypes. The datatypes are set to pointer for every field which need to
// accept default values as valid input.
type JsonCameraInput struct {
    Name   *string               `json:"Name"`
    Ipaddr *string               `json:"Ipaddr"`
    Port   *string               `json:"Port"`
    Desc   *string               `json:"Desc"`
    //Integer status value and '0'(default value) is not a vaid input. so no
    // need to use it as pointer.
    Status dataSet.CameraStatus  `json:"Status"`
    UserId *string               `json:"UserId"`
    Pwd *string                  `json:"Pwd"`
    VideoLenSec uint64           `json:"VideoLenSec"`
}

//Allocate memory to all the string fields that needed for the json structure.
func (jsonCam *JsonCameraInput)AllocateFields() {
    jsonCam.Name = new(string)
    jsonCam.Ipaddr = new(string)
    jsonCam.Port = new(string)
    jsonCam.Desc = new(string)
    jsonCam.UserId = new(string)
    jsonCam.Pwd = new(string)
}

// Read the data from Json structure 'jsonCam' and populate the 'camRowOut'.
// This function leave the'camRowOut' fields untouched when its not set in
// input 'jsonCam'
func (jsonCam *JsonCameraInput)ReadJsonData(camRowOut *dataSet.Camera) {
    if camRowOut == nil {
        // Invalid output structure.
        return
    }
    if jsonCam.Name != nil {
        camRowOut.Name = *jsonCam.Name
    }
    if jsonCam.Ipaddr != nil {
        camRowOut.Ipaddr = *jsonCam.Ipaddr
    }
    if jsonCam.Port != nil {
        camRowOut.Port = *jsonCam.Port
    }
    if jsonCam.Desc != nil {
        camRowOut.Desc = *jsonCam.Desc
    }
    if jsonCam.UserId != nil {
        camRowOut.UserId = *jsonCam.UserId
    }
    if jsonCam.Pwd != nil {
        camRowOut.Pwd = *jsonCam.Pwd
    }
    if jsonCam.Status != 0 {
        camRowOut.Status = jsonCam.Status
    }
    if jsonCam.VideoLenSec != 0 {
        camRowOut.VideoLenSec = jsonCam.VideoLenSec
    }
}
