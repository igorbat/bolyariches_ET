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

package server

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"regexp"

	//"math/rand"
	"net"
	//"regexp"
	"strconv"
	"strings"
	"time"

	serverTypes "fastbot/server/types"
	"fastbot/types"
	e "go-wesnoth/era"
	"go-wesnoth/game"
	"go-wesnoth/scenario"
	"go-wml"
)

type Server struct {
	hostname      string
	port          uint16
	version       string
	username      string
	password      string
	era           e.Era
	game          []byte
	scenarios     serverTypes.ScenarioList
	lastSkip      string // player name
	admins        types.StringList
	observers     types.StringList
	timeout       time.Duration
	err           error
	conn          net.Conn
	disconnecting bool
	sides         serverTypes.SideList
	// Timer-related config
	TimerEnabled  bool
	InitTime      int
	TurnBonus     int
	ReservoirTime int
	ActionBonus   int
	ForceFinish bool
	Ladder *serverTypes.Ladder
	InGame bool
}

var colors = types.StringList{"red", "green", "purple", "orange", "white", "teal"}

func NewServer(hostname string, port uint16, version string, username string,
	password string, era string, scenarios []scenario.Scenario,
	admins types.StringList, timerEnabled bool,
	initTime int, turnBonus int, reservoirTime int, actionBonus int,
	timeout time.Duration, forceFinish bool, ladder *serverTypes.Ladder) *Server {
	s := Server{
		hostname:      hostname,
		port:          port,
		version:       version,
		username:      username,
		password:      password,
		era:           e.Parse(era),
		admins:        admins,
		timeout:       timeout,
		TimerEnabled:  timerEnabled,
		InitTime:      initTime,
		TurnBonus:     turnBonus,
		ReservoirTime: reservoirTime,
		ActionBonus:   actionBonus,
		ForceFinish: forceFinish,
		Ladder: ladder,
		InGame: false,
	}
	var scenarioList serverTypes.ScenarioList
	for _, v := range scenarios {
		scenarioList = append(scenarioList, serverTypes.Scenario{false, v})
	}
	s.scenarios = scenarioList
	s.sides = serverTypes.SideList{
		&serverTypes.Side{Side: 1, Color: "red"},
		&serverTypes.Side{Side: 2, Color: "blue"},
		&serverTypes.Side{Side: 3, Color: "green"},
		&serverTypes.Side{Side: 4, Color: "purple"},
	}
	return &s
}

func (s *Server) Connect() error {
	// Set up a TCP connection
	if s.conn, s.err = net.Dial("tcp", s.hostname+":"+strconv.Itoa(int(s.port))); s.err != nil {
		return s.err
	}
	//s.conn.SetDeadline(time.Now().Add(s.timeout))
	// Init the connection to the server
	s.conn.Write([]byte{0, 0, 0, 0})
	var buffer []byte
	if buffer = s.read(4); s.err != nil {
		return s.err
	}
	fmt.Println("buffer_info", binary.BigEndian.Uint32(buffer))
	// Expects the server to ask for a version, otherwise return an error
	if data := s.receiveData(); bytes.Equal(data, wml.EmptyTag("version").Bytes()) {
		s.sendData((&wml.Tag{"version", wml.Data{"version": s.version}}).Bytes())
	} else {
		return errors.New("Expects the server to request a version, but it doesn't.")
	}
	// Expects the server to require the log in step, otherwise return an error
	{
		rawData := s.receiveData()
		data := wml.ParseData(string(rawData))
		switch {
		case bytes.Equal(rawData, wml.EmptyTag("mustlogin").Bytes()):
			s.sendData((&wml.Tag{"login", wml.Data{"selective_ping": "1", "username": s.username}}).Bytes())
		case data.Contains("redirect"):
			if redirect, ok := data["redirect"].(wml.Data); ok {
				host, okHost := redirect["host"].(string)
				port, okPort := redirect["port"].(string)
				if okHost && okPort {
					portInt, err := strconv.Atoi(port)
					if err == nil {
						s.hostname = host
						s.port = uint16(portInt)
						fmt.Println("REDIRECT " + host + " " + port)
						return s.Connect()
					}
				}
			}
			fallthrough
		default:
			return errors.New("Expects the server to require a log in step, but it doesn't.")
		}
	}
	rawData := s.receiveData()
	data := wml.ParseData(string(rawData))
	switch {
	case data.Contains("error"):
		fmt.Println("ERROR CASE")
		//fmt.Println(data)
		if errorTag, ok := data["error"].(wml.Data); ok {
			if code, ok := errorTag["error_code"].(string); ok {
				switch code {
				case "200":
					if errorTag["password_request"].(string) == "yes" && errorTag["phpbb_encryption"].(string) == "yes" {
						salt := errorTag["salt"].(string)
						qqq := (&wml.Tag{Name: "login", Data: wml.Data{"username": s.username, "password": Sum(s.password, salt)}}).Bytes()
						s.sendData(qqq)
						goto nextCase
					}
				case "105":
					if message, ok := errorTag["message"].(string); ok {
						return errors.New(message)
					} else {
						return errors.New("The nickname is not registered. This server disallows unregistered nicknames.")
					}
				}
			}
		}
		break
	nextCase:
		fallthrough
	case bytes.Equal(rawData, wml.EmptyTag("join_lobby").Bytes()):
		return nil
	default:
		return errors.New("An unknown error occurred")
	}
	return nil
}

func (s *Server) HostGame() {
	if s.InGame {
		return
	}
	s.sides = serverTypes.SideList{
		&serverTypes.Side{Side: 1, Color: "red"},
		&serverTypes.Side{Side: 2, Color: "blue"},
		&serverTypes.Side{Side: 3, Color: "green"},
		&serverTypes.Side{Side: 4, Color: "purple"},
	}
	s.observers = types.StringList{}
	fmt.Println(s.sides[0].Player, s.sides[1].Player, s.sides[2].Player, s.sides[3].Player)

	var path string
	var defines []string
	path = s.scenarios[0].Scenario.Path()
	g := game.NewGame("Isar's Cross Festival: Game " + strconv.Itoa(s.Ladder.GamesCount),
		scenario.FromPath(path, defines),
		s.era,
		s.TimerEnabled, s.InitTime, s.TurnBonus, s.ReservoirTime, s.ActionBonus,
		s.version)
	s.game = g.Bytes()

	s.InGame = true
	s.sendData(
		(&wml.Tag{"create_game", wml.Data{"name": "Isar's Cross Festival: Game " + strconv.Itoa(s.Ladder.GamesCount),
		"password": ""}}).Bytes())
	s.sendData(s.game)
}

func (s *Server) StartGame() {
	s.sendData(wml.EmptyTag("stop_updates").Bytes())
	r, _ := regexp.Compile(`(?Us)\[era\](.*)\[multiplayer_side\]`)
	rDomain, _ := regexp.Compile(`[\t ]*#textdomain ([0-9a-z_-]+)\n`)
	textdomains := rDomain.FindAllSubmatch(r.FindSubmatch(s.game)[1], -1)
	textdomain := string(textdomains[len(textdomains)-1][1])
	rand.Seed(time.Now().UTC().UnixNano())
	qq := []int {0, 1, 2, 3, 4, 5}
	rand.Shuffle(6, func(i, j int) {
		qq[i], qq[j] = qq[j], qq[i]
	})
	data := wml.Data{"scenario_diff": wml.Data{"change_child": wml.Data{
		"index": 0,
		"scenario": wml.Data{"change_child": wml.Multiple{
			wml.Data{"index": 0, "side": insertFaction(s.sides.Side(1), s.era.Factions[qq[0]], textdomain)},
			wml.Data{"index": 1, "side": insertFaction(s.sides.Side(2), s.era.Factions[qq[1]], textdomain)},
			wml.Data{"index": 2, "side": insertFaction(s.sides.Side(3), s.era.Factions[qq[2]], textdomain)},
			wml.Data{"index": 3, "side": insertFaction(s.sides.Side(4), s.era.Factions[qq[3]], textdomain)},
		}},
	},
	}}
	s.sendData(data.Bytes())
	s.sendData(wml.EmptyTag("start_game").Bytes())
}

func (s *Server) LeaveGame() {
	if s.InGame {
		s.sendData(wml.EmptyTag("leave_game").Bytes())
	}
	s.InGame = false
}

func (s *Server) Disconnect() {
	time.Sleep(time.Second * 3)
	s.disconnecting = true
	s.InGame = false
}

func (s *Server) Listen() {
	for {
		if s.disconnecting == true {
			s.conn.Close()
			s.disconnecting = false
			break
		}
		if !s.InGame {
			s.HostGame()
		}
		data := wml.ParseData(string(s.receiveData()))
		switch {
		case data.Contains("name") && data.Contains("side") && s.sides.FreeSlots() > 0:
			if len(data) > 0 {
				fmt.Printf("Received: %q\n", data)
			}
			name := data["name"].(string)
			side, _ := strconv.Atoi(data["side"].(string))
			// if not blacklisted
			if s.sides.HasSide(side) && !s.sides.HasPlayer(name) {
				if s.Ladder.CanPlay(name) {
					s.ChangeSide(side, "insert", wml.Data{"current_player": name, "name": name, "player_id": name})
					s.sides.Side(side).Player = name
					s.sides.Side(side).Ready = true
					if s.sides.MustStart() {
						id, n1, n2, n3, n4 := s.Ladder.ArrangeGame(
							s.sides.Side(1).Player,
							s.sides.Side(2).Player,
							s.sides.Side(3).Player,
							s.sides.Side(4).Player)
						s.sides.Side(1).Player = n1
						s.sides.Side(2).Player = n2
						s.sides.Side(3).Player = n3
						s.sides.Side(4).Player = n4
						s.StartGame()
						s.Ladder.GameStarted(n1, n2, n3, n4)
						time.Sleep(time.Second * 3)
						s.LeaveGame()

						text := "Your game id is " + id + ". Winners: Don't forget to '/whisper \"bot\" won " + id +
							"', when you finish. '/whisper \"bot\" help' for more info"
						s.Whisper(n1, text)
						s.Whisper(n2, text)
						s.Whisper(n3, text)
						s.Whisper(n4, text)
					}
				} else {
					s.Whisper(name, "You can't join: banned or too many unfinished games")
				}
			} else if name != s.username && !s.observers.ContainsValue(name) {
				s.observers = append(s.observers, name)
			}
		case data.Contains("side_drop"):
			side_drop := data["side_drop"].(wml.Data)
			if side_drop.Contains("side_num") {
				side, _ := strconv.Atoi(side_drop["side_num"].(string))
				s.ChangeSide(side, "delete", wml.Data{"current_player": "x", "name": "x", "player_id": "x"})
				sideStruct := s.sides.Side(side)
				sideStruct.Ready = false
				sideStruct.Player = ""
			}
		case data.Contains("observer"):
			observer := data["observer"].(wml.Data)
			if observer.Contains("name") {
				name := observer["name"].(string)
				if name != s.username && !s.observers.ContainsValue(name) {
					s.observers = append(s.observers, name)
				}
			}
		case data.Contains("observer_quit"):
			ObserverQuit := data["observer_quit"].(wml.Data)
			if ObserverQuit.Contains("name") {
				name := ObserverQuit["name"].(string)
				if name != s.username && s.observers.ContainsValue(name) {
					s.observers.DeleteValue(name)
				}
			}
		case data.Contains("leave_game"):
			for _, v := range s.sides {
				v.Player = ""
				v.Ready = false
			}
			s.InGame = false
			s.HostGame()
			for _, v := range s.sides {
				s.ChangeSide(v.Side, "insert", wml.Data{"color": v.Color})
			}
		case data.Contains("whisper"):
			whisper := data["whisper"].(wml.Data)
			if len(data) > 0 {
				fmt.Printf("Received: %q\n", data)
			}
			if whisper.Contains("message") && whisper.Contains("receiver") && whisper.Contains("sender") {
				text := whisper["message"].(string)
				receiver := whisper["receiver"].(string)
				sender := whisper["sender"].(string)
				if receiver == s.username && s.admins.ContainsValue(sender) {
					// Ordinary commands
					command := strings.Fields(text)
					switch {
					case command[0] == "admins" && len(command) == 1:
						s.Whisper(sender, "Admin list: "+strings.Join(s.admins, ", "))
					case command[0] == "remove" && len(command) == 2:
						gameId := types.ParseInt(command[1], -1)
						msg := s.Ladder.GameRemoved(gameId)
						s.Whisper(sender, msg)
					case command[0] == "to_check" && len(command) == 1:
						msg := s.Ladder.ShowContested()
						s.Whisper(sender, "Contested List: " + msg)
					case command[0] =="to_finish" && len(command) == 1:
						msg := s.Ladder.ShowUnfinished()
						s.Whisper(sender, "Unfinished List: " + msg)
					case command[0] == "yellow_card" && len(command) == 2:
						s.Ladder.AddYellowCard(command[1])
						s.Whisper(sender, "Added")
					case command[0] == "green_card" && len(command) == 2:
						s.Ladder.RemoveYellowCard(command[1])
						s.Whisper(sender, "Removed")
					case command[0] == "force_report" && len(command) == 3:
						gameId := types.ParseInt(command[1], -1)
						msg := s.Ladder.GameReported(command[2], gameId)
						s.Whisper(sender, msg)
					case command[0] == "force_contest" && len(command) == 3:
						gameId := types.ParseInt(command[1], -1)
						msg := s.Ladder.GameContested(command[2], gameId)
						s.Whisper(sender, msg)
					case command[0] == "host" && len(command) == 5:
						id,_,_,_,_ := s.Ladder.ArrangeGame(command[1], command[2], command[3], command[4])
						s.Ladder.GameStarted(command[1], command[2], command[3], command[4])
						s.Whisper(sender, "Game started: " + id + " " + command[1] + " " + command[2] + " " + command[3] + " " + command[4])
						s.LeaveGame()
					case command[0] == "host2" && len(command) == 6:
						id,_,_,_,_ := s.Ladder.ArrangeGame(command[1], command[2], command[3], command[4])
						s.Ladder.GameStarted(command[1], command[2], command[3], command[4])
						gameId := types.ParseInt(id, -1)
						msg := s.Ladder.GameReported(command[5], gameId)
						s.Whisper(sender, msg)
						s.LeaveGame()
					case command[0] == "host3" && len(command) == 7:
						id,_,_,_,_ := s.Ladder.ArrangeGame(command[1], command[2], command[3], command[4])
						s.Ladder.GameStarted(command[1], command[2], command[3], command[4])
						gameId := types.ParseInt(id, -1)
						msg := s.Ladder.GameReported(command[5], gameId)
						s.Whisper(sender, msg)
						msg2 := s.Ladder.GameContested(command[6], gameId)
						s.Whisper(sender, msg2)
						s.LeaveGame()
					case command[0] == "host4" && len(command) == 7:
						id,_,_,_,_ := s.Ladder.ArrangeGame(command[1], command[2], command[3], command[4])
						s.Ladder.GameStarted(command[1], command[2], command[3], command[4])
						gameId := types.ParseInt(id, -1)
						msg := s.Ladder.GameReported(command[5], gameId)
						s.Whisper(sender, msg)
						msg2 := s.Ladder.GameContested(command[6], gameId)
						s.Whisper(sender, msg2)
						msg3 := s.Ladder.GameRemoved(gameId)
						s.Whisper(sender, msg3)
						s.LeaveGame()
					case command[0] == "force_finish" && len(command) == 1:
						s.ForceFinish = true
						s.Whisper(sender, "Logging out totally...")
						s.LeaveGame()
						s.Disconnect()
					case command[0] == "stop" && len(command) == 1:
						s.Whisper(sender, "Logging out...")
						s.LeaveGame()
					case command[0] == "dump" && len(command) == 2:
						players, gamehistory := s.Ladder.Dump()
						d1 := []byte("______players____________\n" + players +"_____gamehistory___________\n" + gamehistory)
						ioutil.WriteFile("C:\\Users\\IgorBat\\Desktop\\ET_bots\\dump_" + command[1] + ".txt", d1, 0666)
						s.Whisper(sender, "DUMPED")
					case command[0] == "admin_help" && len(command) == 1:
						s.Whisper(sender, "Command list:\n"+
							"stop - reloads bot, he restarts game\n"+
							"host n1 n2 n3 n4 - host a game in memory with such players order\n"+
							"host2 n1 n2 n3 n4 w1-  host command. report w1 wins\n"+
							"host3 n1 n2 n3 n4 w1 c1- host2 command. contest a game by c1\n"+
							"host4 n1 n2 n3 n4 w1 c1- host3 command. remove game\n"+
							"to_finish - unfinished games\n"+
							"to_check - contested games\n"+
							"force_report id nick - report game id by nick\n" +
							"force_contest id nick - contest game id by nick\n" +
							"admin_help - request command reference\n")
					}
				}
				// for rabotyagi  //////////////////////////////////////////////////////////////////////////////
				if receiver == s.username {
					command := strings.Fields(text)
					switch {
					case command[0] == "won" && len(command) == 2:
						gameId := types.ParseInt(command[1], -1)
						msg := s.Ladder.GameReported(sender, gameId)
						s.Whisper(sender, msg)
					case command[0] == "contest" && len(command) == 2:
						gameId := types.ParseInt(command[1], -1)
						msg := s.Ladder.GameContested(sender, gameId)
						s.Whisper(sender, msg)
					case command[0] == "gameinfo" && len(command) == 2:
						gameId := types.ParseInt(command[1], -1)
						msg := s.Ladder.GetGameInfo(gameId)
						s.Whisper(sender, msg)
					case command[0] == "playerinfo" && len(command) == 2:
						msg := s.Ladder.GetPlayerInfo(command[1])
						s.Whisper(sender, msg)
					case command[0] == "leaderboard" && len(command) == 1:
						msg := s.Ladder.GetLeaderBoard()
						s.Whisper(sender, msg)
					case command[0] == "me" && len(command) == 1:
						msg := s.Ladder.GetPlayerInfo(sender)
						s.Whisper(sender, msg)
					case command[0] == "help" && len(command) == 1:
						s.Whisper(sender, "Command list:\n"+
							"won <game_id> - report the game you have played\n" +
							"contest <game_id> - contest the game you have played\n" +
							"gameinfo <game_id> - info about game\n" +
							"playerinfo <nickname> - info about player\n" +
							"leaderboard - leaderboard\n" +
							"me - info about you\n" +
							"help - request command reference")
					}
				}
			}
		}
	}
}

func (s *Server) ChangeSide(side int, command string, data wml.Data) {
	s.sendData((&wml.Data{"scenario_diff": wml.Data{
		"change_child": wml.Data{"index": 0, "scenario": wml.Data{
			"change_child": wml.Data{"index": side - 1, "side": wml.Data{command: data}},
		}},
	}}).Bytes())
}

func (s *Server) Message(text string) {
	for _, v := range SplitMessage(wml.EscapeString(text)) {
		s.sendData((&wml.Data{"message": wml.Data{"message": v, "room": "this game", "sender": s.username}}).Bytes())
	}
}

//func (s *Server) InGameMessage(text string) {
//	for _, v := range SplitMessage(wml.EscapeString(text)) {
//		s.sendData((&wml.Data{"command": wml.Data{"speak": wml.Data{"message": v, "id": s.username}}}).Bytes())
//	}
//}

func (s *Server) Whisper(receiver string, text string) {
	for _, v := range SplitMessage(wml.EscapeString(text)) {
		s.sendData((&wml.Data{"whisper": wml.Data{"sender": s.username, "receiver": receiver, "message": v}}).Bytes())
	}
}

func (s *Server) Error() error {
	return s.err
}

func (s *Server) receiveData() []byte {
	buffer := s.read(4)
	if len(buffer) < 4 {
		return nil
	}
	size := int(binary.BigEndian.Uint32(buffer))
	reader, _ := gzip.NewReader(bytes.NewBuffer(s.read(size)))
	var result []byte
	if result, s.err = ioutil.ReadAll(reader); s.err != nil {
		return nil
	}
	if s.err = reader.Close(); s.err != nil {
		return nil
	}
	return result
}

func (s *Server) sendData(data []byte) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	gz.Write([]byte(data))
	gz.Close()

	var length int = len(b.Bytes())
	s.conn.Write([]byte{0, 0, byte(length / 256), byte(length % 256)})
	s.conn.Write(b.Bytes())
}

func (s *Server) read(n int) []byte {
	result := []byte{}
	count := 0
	for count < n {
		buffer := make([]byte, n-count)
		var num int
		num, s.err = s.conn.Read(buffer)
		if s.err != nil {
			return nil
		}
		count += num
		result = append(result, buffer[:num]...)
	}
	return result
}

func (s *Server) LoadFromDump() {
	text, _ := ioutil.ReadFile("c:\\Users\\IgorBat\\Desktop\\ET_bots\\to_read.txt")
	s.Ladder.LoadLadderFromFile(string(text))
}
