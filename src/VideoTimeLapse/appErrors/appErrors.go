package appErrors

import (
    "fmt"
)
var (
    INVALID_INPUT = fmt.Errorf("Invalid Input, Cannot process in system")
    DB_ROW_UNIQUE_ERROR = fmt.Errorf("The Row entry is not unique in the system")
    DB_ROW_PRESENT_IN_SYSTEM = fmt.Errorf(`The entry already present in System`)
    DB_ROW_NOT_FOUND = fmt.Errorf("Row entry not found in DB")
)