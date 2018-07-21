package config

import (
    "fmt"
    "flag"
    "os"
    "path/filepath"
    "VideoTimeLapse/logging"
)

type AppConfig struct {
    Ip string
    Port string
    Logfile string
    Loglevel int64
    Dbpath string
    VideoPath string
}

const (
    DEFAULT_LISTEN_PORT = "9000"
    DEFAULT_LOG_LEVEL = logging.Trace
    DEFAULT_PATH = "/tmp/"
    DEFAULT_DB_NAME = "timelapse.db"
)

func (config *AppConfig)printHelp() {
    helpstr := "\n\t Timelapse video from RTSP camera stream" +
        "\n\t An application to create timelapse video from rtsp cameras" +
        "\n\t   USAGE: ./timelapse {ARGS}" +
        "\n\t      ARGS:" +
        "\n\t      -help / -h                          :- Display help and exit." +
        "\n\t      -a <ipAddr> / -ipaddr <ipAddr>      :- Ip address to listen on(Default : localhost)" +
        "\n\t      -p <port> / -port <port>            :- Port to listen on(Default : 9000)" +
        "\n\t      -f <file> / -logfile <file>         :- Optional logfile" +
        "\n\t      -D <path> / -dir <path>             :- Directory for Backend DB & videos(default : /tmp)" +
        "\n\t      -A <db ip> / -dbIp <db ip>          :- Ip address to reach DB server" +
        "\n\t      -P <db port> / -dbPort <dbport>     :- Port to reach DB server" +
        "\n\t      -l <loglevel>/ -loglevel <loglevel> :- loglevel for the application(Default :2)" +
        "\n\t                                             1. Trace" +
        "\n\t                                             2. Info" +
        "\n\t                                             3. Warning" +
        "\n\t                                             4. Error" +
        "\n\n"
    fmt.Print(helpstr)
}

//Read the config from the commandline to the config structure
func (config *AppConfig)InitConfig() {
    var err error
    flag.Usage = config.printHelp
    ipaddrShort := flag.String("a", "127.0.0.1", "Ip address to listen on")
    ipaddrLong := flag.String("ipaddr", "127.0.0.1", "Ip address to listen on")
    portShort := flag.String("p", DEFAULT_LISTEN_PORT, "Port to listen on")
    portLong := flag.String("port", DEFAULT_LISTEN_PORT, "Port to listen on")
    logFileShort := flag.String("f", "", "Optional logfile")
    logFileLong := flag.String("logfile", "", "Optional logfile")
    loglevelShort := flag.Int64("l", DEFAULT_LOG_LEVEL, "loglevel for the application")
    loglevellong := flag.Int64("loglevel", DEFAULT_LOG_LEVEL, "loglevel for the application")
    pathShort := flag.String("D",DEFAULT_PATH, "Backend DB")
    pathLong := flag.String("dir", DEFAULT_PATH, "Backend DB")
    flag.Parse()

    config.Ip = *ipaddrShort
    if config.Ip == "127.0.0.1" {
        config.Ip = *ipaddrLong
    }
    config.Port = *portShort
    if len(config.Port) == 0 {
        config.Port = *portLong
    }
    config.Logfile = *logFileShort
    if len(config.Logfile) == 0 {
        config.Logfile = *logFileLong
    }
    config.Loglevel = *loglevelShort
    if config.Loglevel == DEFAULT_LOG_LEVEL {
        config.Loglevel = *loglevellong
    }
    path := *pathShort
    if *pathShort == DEFAULT_PATH {
        path = *pathLong
    }
    path, err = filepath.Abs(path)
    if err != nil {
        //Failed to get the absoluate path of directory
         path = DEFAULT_PATH

    } else if _, err = os.Stat(path); err != nil {
        //Directory not present in system.
        path = DEFAULT_PATH
    }
    config.Dbpath = path + "/" + DEFAULT_DB_NAME
    config.VideoPath = path
    //TODO :: Need to populate DB server IP and port when needed.
    // For now there is only one db backend is implemented, i.e sqlite.
    // it doesnt need any server ip or port.
}
