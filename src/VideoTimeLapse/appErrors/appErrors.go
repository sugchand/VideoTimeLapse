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