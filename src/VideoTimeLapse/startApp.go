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

package main

import (
    "fmt"
    "os"
    "syscall"
    "os/signal"
    "VideoTimeLapse/config"
    "VideoTimeLapse/sys"
    "VideoTimeLapse/logging"
    "VideoTimeLapse/restAPI"
    "VideoTimeLapse/dataSet/dataSetImpl"
)

func startLoggerService(configObj *config.AppConfig) {
    logger := new(logging.Logging)
    logger.LogInitSingleton(logging.LogLeveltype(configObj.Loglevel),
                            configObj.Logfile)
    logger.Trace("Logging service is started..")
}

func setupDBService(configObj *config.AppConfig) error {
    var err error
    dataObj := dataSetImpl.GetDataSetObj()
    err = dataObj.CreateDBConnection(configObj)
    if err != nil {
        return err
    }
    err = dataObj.CreateDataStoreTables()
    return err
}

func setupRESTService(configObj *config.AppConfig) error {
    resthandler := new(restAPI.RestAPI)
    err := resthandler.RestAPIMainHandler(configObj.Ip, configObj.Port)
    if err != nil {
        return err
    }
    return nil
}

func main() {
    var err error
    configObj := new(config.AppConfig)
    configObj.InitConfig()
    startLoggerService(configObj)
    syncObj := sys.GetAppSyncObj()
    // Wait for all routines to coalesce
    defer syncObj.JoinAllRoutines()

    log := logging.GetLoggerInstance()
    err = setupDBService(configObj)
    if err != nil {
        log.Error("Failed to start the DB service, err: %s", err)
        panic("DB init error")
    }
    err = setupRESTService(configObj)
    if err != nil {
        log.Error("Failed to start REST service")
        panic("Cannot start REST service")
    }
    // Exit the main thread on Ctrl C
    fmt.Println("\n\n\n *** Press Ctrl+C to Exit *** \n\n\n")
    exitsignal := make(chan os.Signal, 1)
    signal.Notify(exitsignal, syscall.SIGINT, syscall.SIGTERM)
    syncObj.AddRoutineInWaitGroup()
    go func() {
        // Blocking the routine for the exit signal.
        <- exitsignal
        syncObj.ExitRoutineInWaitGroup()
        //Send exit signal to all the goroutines.
        syncObj.DestoryAllRoutines()
    }()
}
