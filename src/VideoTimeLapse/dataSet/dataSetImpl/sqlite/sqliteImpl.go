package sqlite

import (
    "fmt"
    "sync"
    "path/filepath"
    "github.com/jmoiron/sqlx"
    _ "github.com/mattn/go-sqlite3"
    "VideoTimeLapse/config"
    "VideoTimeLapse/logging"
    "VideoTimeLapse/dataSet"
)

var dbOnce sync.Once
var sqlObj *SqliteDataStore

type SqliteDataStore struct {
    dblogger *logging.Logging
    DBConn *sqlx.DB
}


//Create a sql connection and store in the datastore object.
// Return '0' on success and errorcode otherwise.
// It is advised to make single handle in entire application as every handle
// uses a connection pool to manage multiple DB requests.
func (sqlds *SqliteDataStore)CreateDBConnection(
                                config *config.AppConfig) error{
    dbDriver := "sqlite3"
    dbFile, err := filepath.Abs(config.Dbpath)
    if err != nil {
        sqlds.dblogger.Error("Failed to open DB file, %s", err.Error())
        return err
    }
    var dbHandle *sqlx.DB
    dbHandle, err = sqlx.Open(dbDriver, dbFile)
    if err != nil {
        sqlds.dblogger.Error("Failed to connect DB %s", err.Error())
        return err
    }
    sqlds.DBConn = dbHandle
    // Serialize the DB access by limiting open connections to 1.
    // This will ensure there are no issues when concurrent threads are
    // accessing the DB file.
    sqlds.DBConn.SetMaxOpenConns(1)
    sqlds.dblogger.Trace("Created sqlite3 DB connection %s", dbFile)
    return nil
}

//Create all the sqlite tables for TimeLapse application.
func (sqlds *SqliteDataStore)CreateDataStoreTables() error {
    if sqlds.DBConn == nil {
        return fmt.Errorf("Null DB connection, cannot create tables")
    }
    camObj := new(sqlCamera)
    camObj.Camera = new(dataSet.Camera)
    camObj.CreateCameraTable(sqlds.DBConn)

    camObj.Name = "camera1!"
    camObj.Ipaddr = "10.0.0.1"
    camObj.Port = "1234"
    camObj.Desc = "wew"
    camObj.Status = 1
    //camObj.InsertCameraEntry(sqlds.DBConn)
    //camObj.InsertCameraEntry(sqlds.DBConn)

    camObj2 := new(sqlCamera)
    camObj2.Camera = new(dataSet.Camera)
    camObj2.Name = "camera"
    camObj2.Ipaddr = "10.0.0.2"
    camObj2.Port = "1234"
    camObj2.Desc = "sdew blha"
    camObj2.Status = 2
    //camObj2.InsertCameraEntry(sqlds.DBConn)
    return nil
}

func (sqlds *SqliteDataStore)AddNewCamera(camera *dataSet.Camera) error {
    camObj := new(sqlCamera)
    camObj.Camera = camera
    return camObj.InsertCameraEntry(sqlds.DBConn)
}

func (sqlds *SqliteDataStore)DeleteCamera(cameraName string) error {
    return nil
}
func (sqlds *SqliteDataStore)UpdateCamera(camera *dataSet.Camera) error {
    return nil
}

func (sqlds *SqliteDataStore)GetCamera(cameraName string) (*dataSet.Camera,
                                       error) {
    var camObj sqlCamera
    camObj.Camera = new(dataSet.Camera)
    camObj.Camera.Name = cameraName
    camObj.GetCameraEntry(sqlds.DBConn)
    return camObj.Camera, nil
}

func(sqlds *SqliteDataStore)GetAllCameras()([]dataSet.Camera, error) {
    camObj := new(sqlCamera)
    dataObj, err := camObj.GetAllCameraEntries(sqlds.DBConn)
    return dataObj, err
}

// Only one SQL datastore object can be present in the system as connection
//pool can be handled in side the database connection itself
func GetsqliteDataStoreObj() *SqliteDataStore {
    //Initialize the global variable.
    dbOnce.Do(func() {
        sqlObj = new(SqliteDataStore)
        sqlObj.dblogger = logging.GetLoggerInstance()
    })
    return sqlObj
}
