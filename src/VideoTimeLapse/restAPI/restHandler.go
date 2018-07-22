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
    "fmt"
    "net/http"
    "github.com/gorilla/handlers"
    "github.com/gorilla/mux"
    "VideoTimeLapse/sys"
    "VideoTimeLapse/logging"

)
type RestAPI struct {
    listenIp string
    listenPort   string
}

// The rest handler thread to manage rest APIs.
//Must be called as a go-routine.
func (handler *RestAPI)RestHandlerThread(addr string,
                 handlerFn http.Handler) {
    log := logging.GetLoggerInstance()
    syncObj := sys.GetAppSyncObj()
    //Go routine to start server listen thread.
    defer syncObj.ExitRoutineInWaitGroup()
    syncObj.AddRoutineInWaitGroup()
    httpObj :=
            &http.Server { Addr: addr,
                Handler: handlerFn}
    log.Trace("Starting REST service at %s:%s", handler.listenIp,
                handler.listenPort)
    go func() {
        fmt.Printf("\n\nApplication starting REST service at %s\n\n",
            handler.listenIp + ":" + handler.listenPort)
        httpObj.ListenAndServe()
        log.Trace("Exiting the REST service.")
    }()
    //Wait for the exit signal.
    syncObj.ReadRestServiceExitSignal()
    httpObj.Close()
    syncObj. ExitRoutineInWaitGroup()
}

//Start the http server with provided routes.
func (handler *RestAPI)startHTTPServer() error {
    var router *mux.Router
    syncObj := sys.GetAppSyncObj()
    allowedOrigins := handlers.AllowedOrigins([]string{"*"})
    allowedMethods := handlers.AllowedMethods(
                []string{"GET", "POST", "DELETE", "PUT", "PATCH"})
    routerObj := new(Routes)
    router = routerObj.NewRouter()
    syncObj.AddRoutineInWaitGroup()
    //Start rest handler thread, that internally call server listen thread.
    go handler.RestHandlerThread(handler.listenIp + ":" + handler.listenPort,
                handlers.CORS(allowedOrigins, allowedMethods)(router))
    return nil
}

// The main handler function for handling http rest request. Normally they are
// running in seperate goroutine.
func (handler *RestAPI)RestAPIMainHandler(listenIp string,
                                              listenPort string) error {
    var err error
    handler.listenIp = listenIp
    handler.listenPort = listenPort
    err = handler.startHTTPServer()
    if err != nil {
        log := logging.GetLoggerInstance()
        log.Error("Exiting Rest handler, err: %s", err)
        return err
    }
    return nil
}

