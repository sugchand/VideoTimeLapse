package RTSPCameraImpl

import (
    "os"
    "time"
    "sync"
    "unsafe"
    "VideoTimeLapse/dataSet"
    "VideoTimeLapse/config"
    "VideoTimeLapse/logging"

)

// RTSP streaming is using FFmpeg libraries to capture the video and create
// MP4 snapshots. At the end of videolen, the snapshots are stitched to make
// single video file.

// #include "videomux.h"
// #include <stdlib.h>
// #cgo LDFLAGS: -lavformat -lavdevice -lavcodec -lavutil
// #cgo CFLAGS: -std=c11
// #cgo pkg-config: libavcodec
import "C"

type RTSPCameraThread struct {
    name string
    ip string
    port string
    uname string
    pwd string
    exitSignal chan bool
    videoPath string
    videoLen uint64 //Length of video to create timelapse.
    videoInterval uint64 //Interval between the video snapshots.
    totOutFrames uint64 //Total number of frames in final timelapse video
    startTime time.Time
    threadLock sync.RWMutex
}

//Initialize the camera thread with all the relevant information.
func (camThread *RTSPCameraThread)InitCameraThread(cam *dataSet.Camera,
                                                   conf *config.AppConfig)(
                                                   error) {
    var err error
    camThread.threadLock.Lock()
    camThread.name = cam.Name
    camThread.ip = cam.Ipaddr
    camThread.port = cam.Port
    camThread.uname = cam.UserId
    camThread.pwd = cam.Pwd
    camThread.exitSignal = make(chan bool)
    camThread.videoPath = conf.VideoPath
    camThread.videoLen = cam.VideoLenSec
    //Create directory for camera streams.
    camThread.videoPath = camThread.videoPath + "/" + camThread.name +
                          camThread.ip
    if _, err = os.Stat(camThread.videoPath); os.IsNotExist(err) {
        err = os.MkdirAll(camThread.videoPath, 0744)
    }
    //TODO :: read it from DB.
    camThread.videoInterval = 60
    if camThread.videoInterval > camThread.videoLen {
        //Interval cannot be more than the total recording duratation.
        camThread.videoInterval = camThread.videoLen
    }
    camThread.threadLock.Unlock()
    return err
}

// RTSP video input.
type Input struct {
    mutex   *sync.RWMutex
    vsInput *C.struct_VSInput
}

func (camThread *RTSPCameraThread)openRTSPInput(inputFormat string,
                                                inputURL string) *Input {
    log := logging.GetLoggerInstance()
    inputFormatC := C.CString(inputFormat)
    inputURLC := C.CString(inputURL)

    input := C.vs_open_input(inputFormatC, inputURLC, C.bool(true))
    if input == nil {
        log.Error("Failed to open RTSP input %s", inputURL)
        C.free(unsafe.Pointer(inputFormatC))
        C.free(unsafe.Pointer(inputURLC))
        return nil
    }
    C.free(unsafe.Pointer(inputFormatC))
    C.free(unsafe.Pointer(inputURLC))

    return &Input{
        mutex:   &sync.RWMutex{},
        vsInput: input,
    }
}

func (camThread *RTSPCameraThread)openMP4Output(outputFile string,
                                   input *Input) *C.struct_VSOutput{
    log := logging.GetLoggerInstance()
    outputFormatC := C.CString("mp4")
    outputURLC := C.CString(outputFile)
    input.mutex.RLock()
    output := C.vs_open_output(outputFormatC, outputURLC, input.vsInput,
                               C.bool(true))
    input.mutex.RUnlock()
    if output == nil {
        log.Error("Failed to open MP4 output file %s", outputFile)
        C.free(unsafe.Pointer(outputFormatC))
        C.free(unsafe.Pointer(outputURLC))
        return nil
    }
    C.free(unsafe.Pointer(outputFormatC))
    C.free(unsafe.Pointer(outputURLC))
    return output
}

func (camThread *RTSPCameraThread)createVideoSnapshot() error {
    log := logging.GetLoggerInstance()
    log.Trace("Creating video snapshot %s", camThread.name)
    return nil
}

func (camThread *RTSPCameraThread)isExitFired() bool {
    select {
        case <-camThread.exitSignal:
            return true
        default:
            return false
    }
}

// Goroutine to execute the camera thread function.
func (camThread *RTSPCameraThread)executeCameraThreadRoutine() error {
    var err error
    var numFramesCopied uint64 = 0
    var defaultSleep uint64
    var elapsedTime uint64 = 0
    defaultSleep = 2 //2 seconds of sleep.
    camThread.threadLock.RLock()
    if camThread.videoInterval < defaultSleep {
        defaultSleep = camThread.videoInterval
    }
    camThread.threadLock.RUnlock()

    var startTime time.Time
    log := logging.GetLoggerInstance()
    log.Trace("Starting the camera thread instance %s", camThread.name)
    camThread.threadLock.Lock()
    camThread.startTime = time.Now()
    camThread.threadLock.Unlock()
    for {
        camThread.threadLock.RLock()
        totOutFrame := camThread.totOutFrames
        vidInterval := camThread.videoInterval
        camThread.threadLock.RUnlock()
        startTime = time.Now()
        //check if exit signal is triggered,
        if camThread.isExitFired() {
            //Exit the loop, as user wanted to kill the thread.
            //create the timelapse video for whatever is present.
            break
        }
        // We are bit lenient here not to take a read lock.
        if numFramesCopied >= totOutFrame {
            //Create the timelapse video from the video snapshots now.
            //Reset the time to start over the timelapse video.
            numFramesCopied = 0
            camThread.threadLock.Lock()
            camThread.startTime = time.Now()
            camThread.threadLock.Unlock()
            elapsedTime = 0
        }
        if elapsedTime >= vidInterval {
            //Only take snapshot at particular interval.
            err = camThread.createVideoSnapshot()
            if err != nil {
                log.Error("Failed to create snapshot for %s err: %s",
                           camThread.name, err)
            }
            elapsedTime = 0
        }
        time.Sleep(time.Duration(defaultSleep))
        elapsedTime = elapsedTime + uint64(time.Now().Sub(startTime))
    }
    log.Trace("Exiting the RTSP camera thread for %s", camThread.name)
    return err
}

// Start the camera timelapse thread as requested. Caller must ensure
//there are no other timelapse threads are running at this time.
// Function to capture camera feed on specified interval. by default it set to
// 1 minute.
func(camThread *RTSPCameraThread)RunCameraThread() error {
    //total number of frames in the final timelapse video.
    camThread.threadLock.Lock()
    camThread.totOutFrames = (camThread.videoLen/camThread.videoInterval) * 60
    camThread.threadLock.Unlock()
    go camThread.executeCameraThreadRoutine()
    return nil
}

//Camera goroutines are never added to mainApp waitgroup. Main app only
// track the camerarunner module. Its responsibility of cameraRunner to exit
// all the running camera threads properly before exiting.
// This function is a blocking call as execute functions check the exit signal
// in specific intervals.
func(camThread *RTSPCameraThread)StopCameraThread() error {
    log := logging.GetLoggerInstance()
    camThread.exitSignal <- true
    log.Trace("Exit signal successfully triggered to %s", camThread.name)
    return nil
}

//Can only operate on existing running camera thread.
func(camThread *RTSPCameraThread)UpdateCameraThread(*dataSet.Camera)(error) {
    camThread.threadLock.Lock()
    camThread.threadLock.Unlock()
    return nil
}
