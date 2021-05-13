// This file is part of Fastbot.
//
// Fastbot is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Fastbot is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Fastbot.  If not, see <https://www.gnu.org/licenses/>.

// fastbot project main.go
package main

import (
	"fastbot/config"
	"fastbot/server"
	serverTypes "fastbot/server/types"
	"fmt"
	"go-wesnoth/era"
	"go-wesnoth/scenario"
	"go-wesnoth/wesnoth"
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
