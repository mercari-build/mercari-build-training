package config

import (
    "log"
    "os"

    "gopkg.in/ini.v1"
)

type ConfigList struct {
    DbName        string
    TestDbName    string
    SQLDriver     string
}

var Config ConfigList

func init() {
    cfg, err := ini.Load("./config/config.ini")
    if err != nil {
        log.Printf("Failed to read file: %v", err)
        os.Exit(1)
    }

    Config = ConfigList{
        DbName:    cfg.Section("db").Key("name").String(),
        TestDbName:    cfg.Section("db").Key("name_test").String(),
        SQLDriver: cfg.Section("db").Key("driver").String(),
    }
}