package RTSPCameraImpl

import (
    "fmt"
    "os"
    "time"
    "sync"
    "sort"
    "unsafe"
    "io/ioutil"
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

// RTSP/mp4 video input.
type Input struct {
    mutex   *sync.RWMutex
    vsInput *C.struct_VSInput
}

type Output struct {
    mutex *sync.RWMutex
    vsOutput *C.struct_VSOutput
}

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

const (
    TIME_DIR_FORMAT = "20060102150405"
    DEFAULT_SNAPSHOT_LEN = 48
)

var rtspOnce sync.Once
//Initialize the camera thread with all the relevant information.
func (camThread *RTSPCameraThread)InitCameraThread(cam *dataSet.Camera,
                                                   conf *config.AppConfig)(
                                                   error) {
    var err error
    //Register the mux,demux and protocols.
    rtspOnce.Do(func() {
        C.vs_setup();
    })
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
    camThread.videoPath = camThread.videoPath + "/" + camThread.name
    if _, err = os.Stat(camThread.videoPath); os.IsNotExist(err) {
        err = os.MkdirAll(camThread.videoPath, 0744)
    }
    //TODO :: read it from DB.
    camThread.videoInterval = 1
    if camThread.videoInterval > camThread.videoLen {
        //Interval cannot be more than the total recording duratation.
        camThread.videoInterval = camThread.videoLen
    }
    camThread.threadLock.Unlock()
    return err
}

func (camThread *RTSPCameraThread)destroyInput(input *Input) {
    if input == nil || input.vsInput == nil{
        return
    }
    input.mutex.Lock()
    defer input.mutex.Unlock()

    if input.vsInput != nil {
        C.vs_destroy_input(input.vsInput)
        input.vsInput = nil
    }
}

func (camThread *RTSPCameraThread)destroyOutput(output *Output,
                                    waitWrite *sync.WaitGroup) {
    if output == nil {
        return
    }
    output.mutex.Lock()
    defer output.mutex.Unlock()
    if output.vsOutput == nil {
        return
    }
    //Wait for all writes to complete before triggering the destroy.
    waitWrite.Wait()
    C.vs_destroy_output(output.vsOutput)
    output.vsOutput = nil
}

func (camThread *RTSPCameraThread)openInput(inputFormat string,
                                            inputURL string) *Input {
    log := logging.GetLoggerInstance()
    inputFormatC := C.CString(inputFormat)
    inputURLC := C.CString(inputURL)

    input := C.vs_open_input(inputFormatC, inputURLC, C.bool(false))
    if input == nil {
        log.Error("Failed to open input %s", inputURL)
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
                                   input *Input) *Output{
    log := logging.GetLoggerInstance()
    outputFormatC := C.CString("mp4")
    outputURLC := C.CString("file:" + outputFile)
    input.mutex.RLock()
    output := C.vs_open_output(outputFormatC, outputURLC, input.vsInput,
                               C.bool(false))
    input.mutex.RUnlock()
    if output == nil {
        log.Error("Failed to open MP4 output file %s", outputFile)
        C.free(unsafe.Pointer(outputFormatC))
        C.free(unsafe.Pointer(outputURLC))
        return nil
    }
    C.free(unsafe.Pointer(outputFormatC))
    C.free(unsafe.Pointer(outputURLC))
    return &Output {
        mutex: &sync.RWMutex{},
        vsOutput: output,
    }
}

//Write input media packet to the output stream.
// Return number of packets written to the output stream
func (camThread *RTSPCameraThread)writePacket(input *Input,
                            output *Output, pktIn *C.AVPacket,
                            waitWrite *sync.WaitGroup) {
    log := logging.GetLoggerInstance()
    input.mutex.RLock()
    output.mutex.Lock()
    defer output.mutex.Unlock()
    waitWrite.Add(1) //Add write operation to waitgroup.
    defer waitWrite.Done()
    if input == nil || output == nil || input.vsInput == nil ||
        output.vsOutput == nil {
        //Try to write a packet if input and output are valid.
        input.mutex.RUnlock()
        return
    }
    pkt := C.av_packet_clone(pktIn)
    writeRes := C.vs_write_packet(input.vsInput, output.vsOutput, pkt,
                                  C.bool(false))
    input.mutex.RUnlock()
    if writeRes == -1 {
        log.Error("Failed to write the packet.")
        C.av_packet_free(&pkt)
        return
    }
    C.av_packet_free(&pkt)
}

//Create a videosnapshot with specific name.
func (camThread *RTSPCameraThread)createVideoSnapshot(fileName string) error {
    var err error
    log := logging.GetLoggerInstance()
    log.Trace("Creating video snapshot %s", camThread.name)
    var input *Input
    camThread.threadLock.RLock()
    videoPath := camThread.videoPath + "/" +
                camThread.startTime.Format(TIME_DIR_FORMAT)
    var url string
    url = "rtsp://" + camThread.ip + ":" + camThread.port + "/"
    if len(camThread.uname) != 0 {
        url = "rtsp://" + camThread.uname + "@" +
                       camThread.ip + ":" + camThread.port + "/"
    }
    if len(camThread.pwd) != 0 {
        //if password is present, the username must be present.
        url = "rtsp://" + camThread.uname + ":" + camThread.pwd + "@" +
                   camThread.ip + ":" + camThread.port + "/"
    }
    input = camThread.openInput("rtsp", url)
    if input == nil || input.vsInput == nil {
        log.Error("Failed to create Input handler %s", camThread.name)
        camThread.threadLock.RUnlock()
        return nil
    }

    camThread.threadLock.RUnlock()

    //Create the output directory if not exists.
    //There is overhead of checking if a directory exists in
    //every snapshot creation. However it make sure there are no
    //error will happen if user delete the folder accidently.
    if _, err = os.Stat(videoPath); os.IsNotExist(err) {
        err = os.MkdirAll(videoPath, 0744)
        if err != nil {
            log.Error("Failed to create directory %s", videoPath)
            camThread.destroyInput(input)
            return err
        }
    }
    videoPath = videoPath + "/" + fileName
    output := camThread.openMP4Output(videoPath, input)
    if output == nil || output.vsOutput == nil {
        log.Trace("Empty Output for %s", videoPath)
        return nil
    }
    //waitgroup for confirm all write complete before destroying the output.
    var waitWrite sync.WaitGroup
    //Read the frames in the loop.
    for numFrames := 0; numFrames<DEFAULT_SNAPSHOT_LEN; {
        var pkt C.AVPacket
        readRes := C.int(0)
        if input == nil || input.vsInput == nil {
            //Cannot read from a null input. return now.
            return nil
        }
        readRes = C.vs_read_packet(input.vsInput, &pkt, C.bool(false))
        if readRes == -1 {
            break
        }
        if readRes == 0 {
            continue
        }
        go camThread.writePacket(input, output, &pkt, &waitWrite)
        numFrames++
    }
    // We dont wanted to block the orignal thread until the destroy finished.
    // destroy will happen only when all the output write completes.
    go camThread.destroyOutput(output, &waitWrite)
    camThread.destroyInput(input)
    log.Trace("Created camera thread snapshot %s", videoPath)
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

//Function to clean up all the snapshot files after the timelapse creation.
func (camThread *RTSPCameraThread)deleteInputSnapshots(dir string,
                                                    files []os.FileInfo) {
    var err error
    var filepath string
    log := logging.GetLoggerInstance()
    for _, file := range files {
        filepath = dir + "/" + file.Name()
        err = os.Remove(filepath)
        if err != nil {
            log.Error("Failed to delete snapshot file %s", filepath)
        }
    }
}

// Function to create timelapse video from snapshots.
// This function go through every snapshot files and stitch together
// to generate final snapshot video
func (camThread *RTSPCameraThread)createTimelapseWithSnapshots(
                                videoPath string) {
    log := logging.GetLoggerInstance()
    files, err := ioutil.ReadDir(videoPath)
    if err != nil {
        log.Error("Failed to read directory, cannot create timelapse, err:%s",
                    err)
        return
    }
    if len(files) == 0 {
        log.Info("Cannot create timelapse video from empty snapshots")
        return
    }
    //Sort files based on its creation time.
    sort.Slice(files, func(i,j int) bool{
    return files[i].ModTime().Before(files[j].ModTime())
    })
    timeLapsePath := videoPath + "/timeLapse"
    if _, err = os.Stat(timeLapsePath); os.IsNotExist(err) {
        err = os.MkdirAll(timeLapsePath, 0744)
        if err != nil {
            log.Error("Failed to create timelapse directory %s",
                        timeLapsePath)
            return
        }
    }

    //Create a list files with all the snapshots.
    var listFile *os.File
    timeLapseList := videoPath + "/timelapseList.txt"
    listFile, err = os.Create(timeLapseList)
    if err != nil {
        log.Error("Failed to create the snapshot list file %s, err :%s",
                    timeLapseList, err)
        return
    }
    for _, snapshot := range files {
       //Create a list file with all the snapshot file names.
       fmt.Fprintf(listFile, "file '%s'\n", snapshot.Name())
    }
    listFile.Close()
    concatInput := camThread.openInput("concat", timeLapseList)
    if concatInput == nil || concatInput.vsInput == nil {
        log.Error("Failed to open concat file %s", timeLapseList)
        return
    }
    timeLapseOutput := camThread.openMP4Output(timeLapsePath +
                                "/timeLapse.mp4",
                                concatInput)
    if timeLapseOutput == nil || timeLapseOutput.vsOutput == nil {
        log.Error("Failed to create timelapse output handler %s",
                        timeLapsePath + "/timeLapse.mp4")
        camThread.destroyInput(concatInput)
    }
    for {
        var pktIn C.AVPacket
        readRes := C.int(0)
        concatInput.mutex.RLock()
        readRes = C.vs_read_packet(concatInput.vsInput, &pktIn,
                                       C.bool(false))
        concatInput.mutex.RUnlock()
        if readRes == -1  {
            break
        }
        if readRes == 0 {
            continue
        }
        pktOut := C.av_packet_clone(&pktIn)
        if pktOut == nil {
            break
        }
        timeLapseOutput.mutex.Lock()
        writeRes := C.vs_write_packet(concatInput.vsInput,
                                     timeLapseOutput.vsOutput, pktOut,
                                     C.bool(false))
        timeLapseOutput.mutex.Unlock()
        if writeRes == -1 {
            log.Error("Failed to write the packet to timeLapse.")
        }
        C.av_packet_free(&pktOut)
    }
    camThread.destroyInput(concatInput)
    if timeLapseOutput != nil && timeLapseOutput.vsOutput != nil {
        timeLapseOutput.mutex.Lock()
        C.vs_destroy_output(timeLapseOutput.vsOutput)
        timeLapseOutput.mutex.Unlock()
    }
    camThread.deleteInputSnapshots(videoPath, files)
}

// Goroutine to execute the camera thread function.
func (camThread *RTSPCameraThread)executeCameraThreadRoutine() error {
    var err error
    var numFramesCopied uint64 = 0
    var defaultSleep uint64
    var elapsedTime uint64 = 0
    defaultSleep = uint64(time.Second.Nanoseconds())//1 second of sleep.
    camThread.threadLock.RLock()
    if (camThread.videoInterval * uint64(time.Second.Nanoseconds())) <
                                        defaultSleep {
        defaultSleep = camThread.videoInterval *
                        uint64(time.Second.Nanoseconds())
    }
    camThread.threadLock.RUnlock()

    var startTime time.Time
    log := logging.GetLoggerInstance()
    log.Trace("Starting the camera thread instance %s", camThread.name)
    camThread.threadLock.Lock()
    camThread.startTime = time.Now()
    camThread.threadLock.Unlock()
    var fileNameInt uint64
    for {
        // We are bit lenient here to read these values onces and use later.
        camThread.threadLock.RLock()
        totOutFrame := camThread.totOutFrames
        vidInterval := camThread.videoInterval
        camThread.threadLock.RUnlock()
        startTime = time.Now()
        //check if exit signal is triggered,
        if camThread.isExitFired() {
            //Exit the loop, as user wanted to kill the thread.
            break
        }
        if numFramesCopied >= totOutFrame {
            //Create the timelapse video from the video snapshots now.
            //Reset the time to start over the timelapse video.
            log.Trace(`Completed the snapshot generation as %d frames are
                       created, Creating timelapse video`, numFramesCopied)
            go camThread.createTimelapseWithSnapshots(
                                   camThread.videoPath + "/" +
                                   camThread.startTime.Format(TIME_DIR_FORMAT))
            numFramesCopied = 0
            camThread.threadLock.Lock()
            camThread.startTime = time.Now()
            camThread.threadLock.Unlock()
            elapsedTime = 0
            fileNameInt = 0
        }
        if elapsedTime >= vidInterval {
            //Only take snapshot at particular interval.
            fileNameInt++
            err = camThread.createVideoSnapshot(fmt.Sprintf("%d.mp4",
                                                fileNameInt))
            if err != nil {
                log.Error("Failed to create snapshot for %s err: %s",
                           camThread.name, err)
            }
            elapsedTime = 0
            //Update the number of frames created so far
            numFramesCopied = numFramesCopied + DEFAULT_SNAPSHOT_LEN
        }
        time.Sleep(time.Duration(defaultSleep))
        elapsedTime = elapsedTime + uint64(time.Now().Sub(startTime).Seconds())
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
    camThread.totOutFrames = (camThread.videoLen/camThread.videoInterval) *
                                DEFAULT_SNAPSHOT_LEN
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
