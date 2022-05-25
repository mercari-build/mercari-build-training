package usersError

import (
	"mercari-build-training-2022/app/models/customErrors"
)

var ErrFindUser = customErrors.AppErr{
    Level:   customErrors.Error,
    Code:    500,
    Message: "Couldn't find user",
}

var ErrPostUser = customErrors.AppErr{
    Level:   customErrors.Error,
    Code:    500,
    Message: "Couldn't post user",
}