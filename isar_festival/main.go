// isar_festival project main.go
package main

import (
	"fmt"
	"go-wesnoth/era"
	"go-wesnoth/scenario"
	"go-wesnoth/wesnoth"
	"isar_festival/config"
	"isar_festival/server"
	serverTypes "isar_festival/server/types"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	config.LoadFromArgs()

	// Apply config to go-wesnoth
	wesnoth.Output = config.TmpDir + "/output"
	era.ErasPath = config.Eras
	scenario.TmpDir = config.TmpDir

	var scenarios []scenario.Scenario
	var __ []string
	scenarios = append(scenarios, scenario.FromPath("4p_Isars_Cross.cfg", __))
	srv := server.NewServer(
		config.Hostname,
		config.Port,
		config.Version,
		"login",
		"password",
		config.Era,
		scenarios,
		config.Admins,
		config.Timer.Enabled,
		config.Timer.InitTime,
		config.Timer.TurnBonus,
		config.Timer.ReservoirTime,
		config.Timer.ActionBonus,
		config.Timeout,
		false,
		serverTypes.LoadLadder("", ""),
		)
	srv.LoadFromDump()
	fmt.Println("Log in started")
	err := srv.Connect()
	check(err)
	for true {
		fmt.Println("Isar hosted")
		time.Sleep(time.Second * 5)
		srv.HostGame()
		srv.Listen()
		if srv.ForceFinish {
			break
		}
	}
}
