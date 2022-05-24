package customErrors

import (
    "errors"
    "fmt"

    "github.com/labstack/echo/v4"
)

// mapのシンタックスシュガー(ある構文を別の記法で記述できるようにしたもの.)
type m map[string]interface{}

// エラーハンドリングの実装
func ErrorHandler(err error, c echo.Context) {
    var appErr *AppErr
    // errors.As()で、代入の可否やUnWrap()を適応してエラーの型を比較
    if errors.As(err, &appErr) {
        switch appErr.Level {
            case Fatal:
                fmt.Errorf("[%s] %d %+v\n", appErr.Level, appErr.Code, appErr.Unwrap())
            case Error:
                fmt.Errorf("[%s] %d %+v\n", appErr.Level, appErr.Code, appErr.Unwrap())
            case Warning:
        }
    } else {
        appErr = ErrUnknown
    }
    c.JSON(appErr.Code, m{"message": appErr.Message})
}

// エラーの型
type AppErr struct {
    Level   ErrLevel
    Code    int
    Message string
    err     error
}

// AppErr構造体のポインタにError関数を定義. これでAppErrはErrorインターフェースを満たす.
func (e *AppErr) Error() string {
    return fmt.Sprintf("[%s] %d: %+v", e.Level, e.Code, e.err)
}

// エラーレベルを表す変数の定義
type ErrLevel string

// エラーレベルの一覧を定数で定義
const (
    Fatal   ErrLevel = "FATAL"
    Error   ErrLevel = "ERROR"
    Warning ErrLevel = "WARNING"
)

var ErrUnknown = &AppErr{
    Level:   Fatal,
    Code:    500,
    Message: "unknown error",
}

func (e *AppErr) Wrap(err error) error {
    e.err = err
    return e
}

func (e *AppErr) Unwrap() error {
    return e.err
}