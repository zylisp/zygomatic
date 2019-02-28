package zcore

import (
	"fmt"
)

var WrongNargs error = fmt.Errorf("wrong number of arguments")

var StopErr error = fmt.Errorf("stop")
