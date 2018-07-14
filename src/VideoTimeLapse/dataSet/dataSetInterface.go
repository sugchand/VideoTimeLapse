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
    "VideoTimeLapse/config"
)

//Dataset Interface that provides the APIs exposed by datastore implementation.
type DataSetInterface interface {
    CreateDBConnection(config *config.AppConfig) error
    // Create all the relevant tables/sessions that are needed for dataset
    //implementation.
    CreateDataStoreTables() error

    //APIs to intract with camera
    AddNewCamera(camera *Camera) error
    DeleteCamera(cameraName string) error
    UpdateCamera(camera *Camera) error
    GetCamera(cameraName string) (*Camera, error)
    GetAllCameras()([]Camera, error)
}
