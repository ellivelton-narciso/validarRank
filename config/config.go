package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type UserStruct struct {
	Host        string `json:"host"`
	User        string `json:"user"`
	Pass        string `json:"pass"`
	Port        string `json:"port"`
	Dbname      string `json:"dbname"`
	AlertasDisc string `json:"alertasDisc"`
}

var (
	Host        string
	User        string
	Pass        string
	Port        string
	DBname      string
	AlertasDisc string
	UserConfig  UserStruct
)

func ReadFile() {
	user, err := ioutil.ReadFile("config.json")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	err = json.Unmarshal(user, &UserConfig)

	Host = UserConfig.Host
	User = UserConfig.User
	Pass = UserConfig.Pass
	Port = UserConfig.Port
	DBname = UserConfig.Dbname
	AlertasDisc = UserConfig.AlertasDisc

}
