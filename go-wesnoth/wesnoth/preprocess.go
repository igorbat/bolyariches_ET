// This file is part of Go Wesnoth.
//
// Go Wesnoth is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Go Wesnoth is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Go Wesnoth.  If not, see <https://www.gnu.org/licenses/>.

package wesnoth

import (
	"io/ioutil"
	"os"
	"fmt"
)

var (
	Wesnoth = "/usr/bin/wesnoth"
	Output  = os.TempDir() + "/go-wesnoth/output"
)

func Preprocess(typee string, defines []string) []byte {
	//defines = append(defines, "MULTIPLAYER")
	//if _, err := os.Stat(Output); os.IsNotExist(err) {
	//	os.MkdirAll(Output, 0755)
	//}
	//cmd := exec.Command(
	//	Wesnoth,
	//	"-p",
	//	filePath,
	//	Output,
	//	"--preprocess-defines=MULTIPLAYER"+strings.Join(defines, ","),
	//)
	//cmd.Run()
	var pathh string
	if (typee == "eras") {
		pathh = "c:\\Users\\IgorBat\\Desktop\\ET_bots\\eras.cfg"
	}
	if (typee == "units"){
		pathh = "c:\\Users\\IgorBat\\Desktop\\ET_bots\\units.cfg"
	}
	if (typee == "isar") {
		pathh = "c:\\Users\\IgorBat\\Desktop\\ET_bots\\4p_Isars_Cross.cfg"
	}
	result, err := ioutil.ReadFile(pathh)
	fmt.Println("In preprocess " + typee + " captured ", err)
	return result
}
