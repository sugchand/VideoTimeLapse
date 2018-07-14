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
    "github.com/gorilla/mux"
    "VideoTimeLapse/logging"
    "VideoTimeLapse/dataSet/dataSetImpl"
)

// Route defines a route
type routeEntry struct {
    Name        string
    Method      string
    Pattern     string
    HandlerFunc http.HandlerFunc
}

type Routes struct {
    entries []routeEntry
    controller *controller
}

func (routeObj *Routes) CreateAllRoutes() {
    routeObj.entries = make([]routeEntry, 9)
    routeObj.entries[0] = routeEntry{
                            "getAllCameras",
                            "GET",
                            "/cameras",
                            routeObj.controller.getAllCameras}
    routeObj.entries[1] = routeEntry{
                            "getCamera",
                            "GET",
                            "/cameras/{camera-name}",
                            routeObj.controller.getCamera}
    routeObj.entries[2] = routeEntry{
                            "CreateCamera",
                            "POST",
                            "/cameras",
                            routeObj.controller.createCamera}
    routeObj.entries[3] = routeEntry{
                            "updateCamera",
                            "PATCH",
                            "/cameras/{camera-name}",
                            routeObj.controller.updateCamera}
    routeObj.entries[4] = routeEntry{
                            "deleteCamera",
                            "DELETE",
                            "/cameras/{camera-name}",
                            routeObj.controller.deleteCamera}
    //Video REST API routes
    routeObj.entries[5] = routeEntry{
                            "getVideos",
                            "GET",
                            "/cameras/{camera-name}/videos",
                            routeObj.controller.getVideos}
    routeObj.entries[6] = routeEntry{
                            "deleteVideos",
                            "DELETE",
                            "/cameras/{camera-name}/videos",
                            routeObj.controller.deleteVideos}
    routeObj.entries[7] = routeEntry{
                            "getVideo",
                            "GET",
                            "/cameras/{camera-name}/videos/{video-name}",
                            routeObj.controller.getVideo}
    routeObj.entries[8] = routeEntry{
                            "deleteVideo",
                            "DELETE",
                            "/cameras/{camera-name}/videos/{video-name}",
                            routeObj.controller.deleteVideo}
}

// NewRouter function configures a new router to the API
func (routeObj *Routes)NewRouter() *mux.Router {
    log := logging.GetLoggerInstance()
    router := mux.NewRouter().StrictSlash(true)
    routeObj.CreateAllRoutes()
    for _, route := range routeObj.entries {
        var handler http.Handler
        handler = route.HandlerFunc
        router.
         Methods(route.Method).
         Path(route.Pattern).
         Name(route.Name).
         Handler(handler)
        log.Trace("Created route for %s", route.Name)
    }
    routeObj.controller = new(controller)
    routeObj.controller.dataObj = dataSetImpl.GetDataSetObj()
    return router
}
