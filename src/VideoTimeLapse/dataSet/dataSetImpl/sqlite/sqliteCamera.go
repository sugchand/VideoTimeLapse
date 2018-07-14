package sqlite

import (
    "fmt"
    "github.com/jmoiron/sqlx"
    "VideoTimeLapse/dataSet"
    "VideoTimeLapse/logging"
)

type Camera struct {
    name   string
    ipaddr string
    port   string
    desc   string
    status dataSet.CameraStatus
}

const (
    CAMERA_TABLE = "camera"
    CAMERA_FIELD_NAME = "name"
    CAMERA_FIELD_IPADDR = "ipaddr"
    CAMERA_FIELD_PORT = "port"
    CAMERA_FIELD_DESC = "desc"
    CAMERA_FIELD_STATUS = "status"
    CAMERA_FIELD_USERID = "userid"
    CAMERA_FIELD_PWD = "pwd"
)

var (
    cameraSchema = fmt.Sprintf(
                `CREATE TABLE IF NOT EXISTS %s (%s TEXT PRIMARY KEY,
                 %s TEXT NOT NULL,
                 %s TEXT NOT NULL,
                 %s TEXT,
                 %s INTEGER,
                 %s TEXT,
                 %s TEXT)`,
                 CAMERA_TABLE,
                 CAMERA_FIELD_NAME,
                 CAMERA_FIELD_IPADDR,
                 CAMERA_FIELD_PORT,
                 CAMERA_FIELD_DESC,
                 CAMERA_FIELD_STATUS,
                 CAMERA_FIELD_USERID,
                 CAMERA_FIELD_PWD)
    //Create a role entry in table roles
    cameraCreate = fmt.Sprintf(`INSERT INTO %s
                                (%s, %s, %s, %s, %s, %s, %s)
                                VALUES (?, ?, ?, ?,?,?,?)`,
                                CAMERA_TABLE,
                                CAMERA_FIELD_NAME,
                                CAMERA_FIELD_IPADDR,
                                CAMERA_FIELD_PORT,
                                CAMERA_FIELD_DESC,
                                CAMERA_FIELD_STATUS,
                                CAMERA_FIELD_USERID,
                                CAMERA_FIELD_PWD)

    cameraGet = fmt.Sprintf("SELECT * FROM %s WHERE %s=(?)",
                            CAMERA_TABLE,
                            CAMERA_FIELD_NAME)
    cameraGetAll = fmt.Sprintf("SELECT * FROM %s", CAMERA_TABLE)
    cameraDelete = fmt.Sprintf("DELETE FROM %s WHERE %s=(?)",
                                CAMERA_TABLE, CAMERA_FIELD_NAME)

)

type sqlCamera struct {
    *dataSet.Camera
}

func(camObj *sqlCamera)CreateCameraTable(conn *sqlx.DB) error {
    var err error
    log := logging.GetLoggerInstance()
    _, err = conn.Exec(cameraSchema)
    if err != nil {
        log.Error("Failed to create Camera table %s", err)
        return err
    }
    return nil
}

func (camObj *sqlCamera)GetAllCameraEntries(conn *sqlx.DB) ([]dataSet.Camera, error) {
    var err error
    log := logging.GetLoggerInstance()
    rows := []dataSet.Camera{}
    err = conn.Select(&rows, cameraGetAll)
    if err != nil {
        log.Error("Failed to gt the rows from camera table")
    }
    return rows, err
}

//Return the camera entry and number of entries in the DB.
// The structure will overwritten by the DB row.
func(camObj *sqlCamera)GetCameraEntry(conn *sqlx.DB) (int, error) {
    var err error
    log := logging.GetLoggerInstance()
    rows := []dataSet.Camera{}
    err = conn.Select(&rows, cameraGet, camObj.Name)
    if err != nil {
        log.Error("Failed to get the row for %s", camObj.Name)
        return 0, err
    }
    if len(rows) == 0 {
        var zeroCamera dataSet.Camera
        camObj.Camera = &zeroCamera
        return 0, nil
    }
    camObj.Camera = &rows[0]
    return len(rows), nil
}

func(camObj *sqlCamera)InsertCameraEntry(conn *sqlx.DB) error {
    var err error
    var numEntries int
    log := logging.GetLoggerInstance()
    isvalid, _ := camObj.Camera.IsCameraStatusValid()
    if !isvalid {
        log.Error("Cannot create Camera entry %s, Invalid status %d",
                    camObj.Status, camObj.Name)
        return fmt.Errorf("Cannot create camera entry %s, invalid status",
                            camObj.Name)
    }
    dbRow := new(sqlCamera)
    *dbRow = *camObj
    numEntries, err = dbRow.GetCameraEntry(conn)
    if  err != nil {
        log.Error("Failed to get the camera record for %s", camObj.Name)
        return err
    }
    if numEntries != 0 {
        log.Info("Camera entry :%s is present in the DB", camObj.Name)
        return nil
    }
    _, err = conn.Exec(cameraCreate, camObj.Name, camObj.Ipaddr, camObj.Port,
                        camObj.Desc, camObj.Status, camObj.UserId, camObj.Pwd)
    if err != nil {
        log.Error("Failed to create the camera record %s, err :%s",
                            camObj.Name, err)
        return err
    }
    return nil
}

func(camObj *sqlCamera)DeleteCameraEntry(conn *sqlx.DB) error {
    var err error
    log := logging.GetLoggerInstance()
    _, err = conn.Exec(cameraDelete, camObj.Name)
    if err != nil {
        log.Error("Failed to delete camera entry err: %s", err)
        return err
    }
    return nil
}

