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

package config

import (
	"os"
	"time"

	"fastbot/types"
)

// Default Bot's parameters
var (
	// Accessed from the outside
	// Hostname        = "127.0.0.1"
	Hostname        = "95.217.86.148"//"server.wesnoth.org" //"127.0.0.1"9 5.217.86.148 14999
	Port     uint16 = 15000 //15000
	Version         = "1.14.7"
	Era             = "default"
	Timer           = TimerConfig{true, 210, 210, 210, 0}
	Admins          = types.StringList{"igorbat99", "dwarftough", "igorbat99_2"}
	BaseDir         = "C:\\Program Files (x86)\\Battle for Wesnoth 1.14.7\\data\\multiplayer\\scenarios"

	// Not need to be accessed from the outside
	scenarios = []ScenarioConfig{ScenarioConfig{Path: "C:\\Program Files (x86)\\Battle for Wesnoth 1.14.7\\data\\multiplayer\\scenarios\\4p_Isars_Cross.cfg"}}
	//Games           = []GameConfig{Scenarios: ScenarioConfig{Path: "C:\\Program Files (x86)\\Battle for Wesnoth 1.14.7\\data\\multiplayer\\scenarios\\4p_Isars_Cross.cfg"}}
	// Game distro related confs and timeouts (accessed from the outside)
	//Wesnoth = "/usr/bin/wesnoth"
	//Eras    = "/usr/share/wesnoth/data/multiplayer/eras.cfg"
	//Units   = "/usr/share/wesnoth/data/core/units.cfg"

	Wesnoth = "C:\\Program Files (x86)\\Battle for Wesnoth 1.14.7\\wenoth.exe"
	Eras    = "C:\\Program Files (x86)\\Battle for Wesnoth 1.14.7\\data\\multiplayer\\eras.cfg"
	Units   = "C:\\Program Files (x86)\\Battle for Wesnoth 1.14.7\\data\\core\\units.cfg"

	TmpDir  = os.TempDir() + "/fastbot"
	Timeout = time.Second * 30
)

type GameConfig struct {
	Title         string
	Players       []string
	PickingPlayer string
	Scenarios     []ScenarioConfig
}

type ScenarioConfig struct {
	Path    string
	Defines []string
}

type TimerConfig struct {
	Enabled       bool
	InitTime      int
	TurnBonus     int
	ReservoirTime int
	ActionBonus   int
}

func LoadFromArgs() {
}
