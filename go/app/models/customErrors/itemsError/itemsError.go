package itemsError

import (
	"mercari-build-training-2022/app/models/customErrors"
)

var ErrorHandler = customErrors.ErrorHandler

var ErrGetItems = customErrors.AppErr{
    Level:   customErrors.Error,
    Code:    500,
    Message: "Couldn't get items",
}

var ErrPostItem = customErrors.AppErr{
    Level:   customErrors.Error,
    Code:    500,
    Message: "Couldn't post item",
}