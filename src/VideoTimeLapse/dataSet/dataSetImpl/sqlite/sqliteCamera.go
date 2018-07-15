package sqlite

import (
    "fmt"
    "strconv"
    "github.com/jmoiron/sqlx"
    "VideoTimeLapse/dataSet"
    "VideoTimeLapse/logging"
    "VideoTimeLapse/appErrors"
)

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
                                VALUES (?, ?, ?, ?, ?, ?, ?)`,
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
    cameraGetOnIpPort = fmt.Sprintf(`SELECT * FROM %s WHERE %s=(?) AND
                                        %s=(?)`,
                                    CAMERA_TABLE,
                                    CAMERA_FIELD_IPADDR,
                                    CAMERA_FIELD_PORT)
    cameraGetAll = fmt.Sprintf("SELECT * FROM %s", CAMERA_TABLE)
    cameraUpdate = fmt.Sprintf(`UPDATE %s SET %s=(?),%s=(?),%s=(?),
                                              %s=(?),%s=(?),%s=(?)
                                              WHERE %s=(?)`,
                                              CAMERA_TABLE,
                                              CAMERA_FIELD_IPADDR,
                                              CAMERA_FIELD_PORT,
                                              CAMERA_FIELD_DESC,
                                              CAMERA_FIELD_STATUS,
                                              CAMERA_FIELD_USERID,
                                              CAMERA_FIELD_PWD,
                                              CAMERA_FIELD_NAME)
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

//Find camera entries in the table that matches on specific IP and port.
//DB can only hold one entry with the specific IP and port address.
// Along with camera name, the IP + port must be unique in the DB table.
// @Returns
// The list of rows that has same IP and port. It must be one entry in the list.
func(camObj *sqlCamera)GetCameraEntryonIpPort(conn *sqlx.DB)([]dataSet.Camera,
                                              error) {
    var err error
    log := logging.GetLoggerInstance()
    rows := []dataSet.Camera{}
    if len(camObj.Ipaddr) == 0 || len(camObj.Port) == 0 {
        log.Error("Null IP/Port. Cannot find the entry in the DB")
        return nil, appErrors.INVALID_INPUT
    }
    //Check if port is valid integer or not.
    if _, err = strconv.Atoi(camObj.Port); err != nil {
        log.Error("Invalid Port, Cannot create the camera entry for %s",
                    camObj.Name)
        return nil, err
    }

    err = conn.Select(&rows, cameraGetOnIpPort, camObj.Ipaddr, camObj.Port)
    if err != nil {
        log.Error("Failed to get the row for ip :%s, port :%s",
                        camObj.Ipaddr, camObj.Port)
        return nil, err
    }
    return rows, nil
}

//Return the camera entry and number of entries in the DB.
// The structure will overwritten by the DB row.
// @Returns
//    Number of entries that present in tables. Possible values can be 1 or 0
//    Error value if there are any error happened to read the entry.
func(camObj *sqlCamera)GetCameraEntry(conn *sqlx.DB) (*dataSet.Camera, error) {
    var err error
    log := logging.GetLoggerInstance()
    rows := []dataSet.Camera{}
    err = conn.Select(&rows, cameraGet, camObj.Name)
    if err != nil {
        log.Error("Failed to get the row for %s", camObj.Name)
        return nil, err
    }
    if len(rows) > 1 {
        return &rows[0], appErrors.DB_ROW_UNIQUE_ERROR
    }
    if len(rows) == 0 {
        return nil, appErrors.DB_ROW_NOT_FOUND
    }
    return &rows[0], nil
}

func(camObj *sqlCamera)InsertCameraEntry(conn *sqlx.DB) error {
    var err error

    log := logging.GetLoggerInstance()
    isvalid, _ := camObj.Camera.IsCameraStatusValid()
    if !isvalid {
        log.Error("Cannot create Camera entry %s, Invalid status %d",
                    camObj.Name, camObj.Status)
        return appErrors.INVALID_INPUT
    }
    var row *dataSet.Camera
    row, err = camObj.GetCameraEntry(conn)
    if  err != nil  && err != appErrors.DB_ROW_NOT_FOUND {
        log.Error("Failed to get the camera record for %s", camObj.Name)
        return err
    }
    if row != nil {
        //We found an row with same name,
        log.Error("Cannot insert DB row %s, as its present in system",
                        camObj.Name)
        return appErrors.DB_ROW_PRESENT_IN_SYSTEM
    }
    rows := []dataSet.Camera{}
    rows, err = camObj.GetCameraEntryonIpPort(conn)
    if err != nil {
        log.Error("Failed to insert cameray entry %s", camObj.Name)
        return err
    }
    if len(rows) != 0 {
        log.Error("Camera entry already present with same IP and port")
        return appErrors.DB_ROW_PRESENT_IN_SYSTEM
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

// Update the camera entry, Caller must confirm if the record is present or not
// before trying to update it.
func(camObj *sqlCamera)UpdateCameraEntry(conn *sqlx.DB) error {
    var err error
    log := logging.GetLoggerInstance()
    _, err = conn.Exec(cameraUpdate, camObj.Camera.Ipaddr, camObj.Camera.Port,
                        camObj.Camera.Desc, camObj.Camera.Status,
                        camObj.Camera.UserId, camObj.Camera.Pwd,
                        camObj.Name)
    if err != nil {
        log.Error("Failed to update the camera record err :%s", err)
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
