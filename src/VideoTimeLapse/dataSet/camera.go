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

package dataSet

import (
    "math"
)

const (
    CAMERA_OFF = iota + 1
    CAMERA_ON
    CAMERA_STREAMING
    CAMERA_DELETED
)
//Can be one of the values in the above.
type CameraStatus uint64

//Type of camera protocol.
const (
    CAMERA_TYPE_RTSP = iota + 1
)

//Default value for some of camera parameters.
const (
    //default recording time for a camera.
    CAMERA_DEFAULT_TIMELAPSE_SEC = 3600 //(1 Hr)
    CAMERA_DEFAULT_SNAPSHOT_INTERVAL = 60 //60 seconds
)
//Structure to hold all the information for the camera.
//Must update JsonCameraInput when updating this structure.
type Camera struct {
    Name   string        `json:"Name"`
    Ipaddr string        `json:"Ipaddr"`
    Port   string        `json:"Port"`
    Desc   string        `json:"Desc"`
    Status CameraStatus  `json:"Status"`
    Type   uint64        `json:"Type"`
    UserId string        `json:"UserId"`
    Pwd string           `json:"Pwd"`
    //Total video recording time for timelapse video.
    VideoLenSec uint64   `json:"VideoLenSec"`
    // Interval between videosnapshots
    SnapInterval uint64   `json:"VideoSnapInterval"`
}

func (camObj *Camera) IsCameraStatusValid() (bool, error) {
    if camObj.Status < CAMERA_OFF ||
        camObj.Status > CAMERA_DELETED {
        return false, nil
    }
    return true, nil
}

//Check if Video Param is va
func (camObj *Camera) IsVideoLenValid() (bool) {
    //Minimum of 2 seconds
    if camObj.VideoLenSec < 120 || camObj.VideoLenSec > math.MaxInt32 {
        return false
    }
    return true
}

//Videolen must be set when checking the snapshot length
func (camObj *Camera) IsSnapshotLenValid() (bool) {
    if camObj.SnapInterval < 1 || camObj.SnapInterval >= camObj.VideoLenSec {
        return false
    }
    return true
}
