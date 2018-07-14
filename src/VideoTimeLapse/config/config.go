package config

import (
    "fmt"
    "flag"
    "VideoTimeLapse/logging"
)

type AppConfig struct {
    Ip string
    Port string
    Logfile string
    Loglevel int64
    Dbpath string
}

const (
    DEFAULT_LISTEN_PORT = "9000"
    DEFAULT_LOG_LEVEL = logging.Trace
    DEFAULT_DB_NAME = "/tmp/timelapse.db"
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
        "\n\t      -d <db> / -db <db>                  : Backend DB(default : /tmp/timelapse.db)" +
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
    flag.Usage = config.printHelp
    ipaddrShort := flag.String("a", "127.0.0.1", "Ip address to listen on")
    ipaddrLong := flag.String("ipaddr", "127.0.0.1", "Ip address to listen on")
    portShort := flag.String("p", DEFAULT_LISTEN_PORT, "Port to listen on")
    portLong := flag.String("port", DEFAULT_LISTEN_PORT, "Port to listen on")
    logFileShort := flag.String("f", "", "Optional logfile")
    logFileLong := flag.String("logfile", "", "Optional logfile")
    loglevelShort := flag.Int64("l", DEFAULT_LOG_LEVEL, "loglevel for the application")
    loglevellong := flag.Int64("loglevel", DEFAULT_LOG_LEVEL, "loglevel for the application")
    dbPathShort := flag.String("d",DEFAULT_DB_NAME, "Backend DB")
    dbPathLong := flag.String("db", DEFAULT_DB_NAME, "Backend DB")
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
    config.Dbpath = *dbPathShort
    if *dbPathShort == DEFAULT_DB_NAME {
        config.Dbpath = *dbPathLong
    }

    //TODO :: Need to populate DB server IP and port when needed.
    // For now there is only one db backend is implemented, i.e sqlite.
    // it doesnt need any server ip or port.
}
