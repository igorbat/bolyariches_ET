package types

import (
	"fastbot/types"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Ladder struct {
	Players map[string]*IsarPlayer
	Games map[int]*SavedGame
	GamesCount int
}

func (L *Ladder) Dump() (string, string) {
	var playersDump, gamesDump = "", ""

	var checked []IntStringPair

	kek := func(p1, p2 *IntStringPair) bool {
		return p1.kk > p2.kk || (p1.kk == p2.kk && p1.vv <= p2.vv)
	}

	for k := range L.Players {
		checked = append(checked, IntStringPair{L.Players[k].Raiting, k})
	}

	By(kek).Sort(checked)

	for _, r_player := range checked {
		playersDump += r_player.vv + " " + strconv.Itoa(r_player.kk) + "\n"
	}
	keys := make([]int, 0, len(L.Games))
	for k := range L.Games {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	for _, k := range keys {
		gamesDump += L.Games[k].Dump()
	}
	return playersDump, gamesDump
}

func LoadLadder(_ string, _ string) *Ladder {
	// TODO
	// 1 + 2 = 5
	l := Ladder{Games: make(map[int]*SavedGame), Players: make(map[string]*IsarPlayer), GamesCount: 77}
	return &l
}

func (L *Ladder) LoadLadderFromFile(text string)  {
	strs := strings.Split(text, "\n")
	for _, str := range strs {
		if len(str) == 0 {
			continue
		}
		words := strings.Fields(str)
		if words[0] == "host" {
			id,_,_,_,_ := L.ArrangeGame(words[1], words[2], words[3], words[4])
			gameId := types.ParseInt(id, -1)
			L.GameStarted(words[1], words[2], words[3], words[4])
			fmt.Println("this game is hosted", gameId)
		} else if words[0] == "host2" {
			id,_,_,_,_ := L.ArrangeGame(words[1], words[2], words[3], words[4])
			L.GameStarted(words[1], words[2], words[3], words[4])
			gameId := types.ParseInt(id, -1)
			msg := L.GameReported(words[5], gameId)
			fmt.Println("this game is reported", gameId, msg)
		} else if words[0] == "host3" {
			id,_,_,_,_ := L.ArrangeGame(words[1], words[2], words[3], words[4])
			L.GameStarted(words[1], words[2], words[3], words[4])
			gameId := types.ParseInt(id, -1)
			msg := L.GameReported(words[5], gameId)
			msg1 := L.GameContested(words[6], gameId)
			fmt.Println("this game is contested", gameId, msg, msg1)
		} else if words[0] == "host4" {
			fmt.Println("this game is removed")
			id,_,_,_,_ := L.ArrangeGame(words[1], words[2], words[3], words[4])
			L.GameStarted(words[1], words[2], words[3], words[4])
			gameId := types.ParseInt(id, -1)
			msg := L.GameReported(words[5], gameId)
			msg1 := L.GameContested(words[6], gameId)
			msg2 := L.GameRemoved(gameId)
			fmt.Println("this game is contested", gameId, msg, msg1, msg2)
		} else {
			fmt.Println("WTF is this??", words[:])
		}
	}
}

func (L *Ladder) ShowContested() string {
	var res []int
	for gameId, Game := range L.Games {
		if Game.Contested && !Game.Revoked {
			res = append(res, gameId)
		}
	}
	return dump_slice(-1, res)
}

func (L *Ladder) ShowUnfinished() string {
	var res []int
	for gameId, Game := range L.Games {
		if !Game.Finished {
			res = append(res, gameId)
		}
	}
	return dump_slice(-1, res)
}

func (L* Ladder) GameStarted(n1, n2, n3, n4 string) {
	L.Games[L.GamesCount] = newGame(L.GamesCount, n1, n2, n3, n4,
		L.Players[n1].Raiting, L.Players[n2].Raiting,L.Players[n3].Raiting,L.Players[n4].Raiting)
	L.Players[n1].StartGame(L.GamesCount)
	L.Players[n2].StartGame(L.GamesCount)
	L.Players[n3].StartGame(L.GamesCount)
	L.Players[n4].StartGame(L.GamesCount)
	L.GamesCount += 1
}

func (L* Ladder) GetOrCreateRaiting(nickname string) int {
	v, found := L.Players[nickname]
	if !found {
		L.Players[nickname] = newIsarPlayer(nickname)
		return 1500
	}
	return v.Raiting
}

func (L *Ladder) ArrangeGame(n1,n2,n3,n4 string) (string, string, string, string, string) {
	var r1, r2, r3, r4 int
	r1 = L.GetOrCreateRaiting(n1)
	r2 = L.GetOrCreateRaiting(n2)
	r3 = L.GetOrCreateRaiting(n3)
	r4 = L.GetOrCreateRaiting(n4)
	var checked = []IntStringPair{
		{kk: r1, vv: n1},
		{kk: r2, vv: n2},
		{kk: r3, vv: n3},
		{kk: r4, vv: n4},
	}

	kek := func(p1, p2 *IntStringPair) bool {
		return p1.kk < p2.kk || (p1.kk == p2.kk && p1.vv <= p2.vv)
	}
	By(kek).Sort(checked)
	return strconv.Itoa(L.GamesCount), checked[0].vv, checked[1].vv, checked[2].vv, checked[3].vv
}

func (L *Ladder) GameReported(nickname string, gameId int) string {
	v, found := L.Games[gameId]
	if !found {
		return "No such game"
	}
	f, side := v.wasPlayer(nickname)
	if !f {
		return "You're not player of that game"
	}
	if v.Finished {
		return "Game already reported by " + v.WhoReported
	}
	L.Games[gameId].WhoReported = nickname
	r1, r2, r3, r4 := L.Games[gameId].FinishGame(side)
	fmt.Println("Game finished. Elo changes for " +
		v.Nick1 + " " + strconv.Itoa(r1) + " " +
		v.Nick2 + " " + strconv.Itoa(r2) + " " +
		v.Nick3 + " " + strconv.Itoa(r3) + " " +
		v.Nick4 + " " + strconv.Itoa(r4))
	L.Players[v.Nick1].FinishGame(gameId, r1, false)
	L.Players[v.Nick2].FinishGame(gameId, r2, false)
	L.Players[v.Nick3].FinishGame(gameId, r3, false)
	L.Players[v.Nick4].FinishGame(gameId, r4, false)
	return "Reported"
}

func (L *Ladder) GameRemoved(gameId int) string {
	v, found := L.Games[gameId]
	if !found {
		return "no such game"
	}
	if !L.Games[gameId].Finished {
		return "need to finish before"
	}
	L.Games[gameId].Revoked = true
	r1, r2, r3, r4 := L.Games[gameId].FinishGame(L.Games[gameId].winner_side)
	fmt.Println("Game removed. Elo reverted for " +
		v.Nick1 + " " + strconv.Itoa(r1) + " " +
		v.Nick2 + " " + strconv.Itoa(r2) + " " +
		v.Nick3 + " " + strconv.Itoa(r3) + " " +
		v.Nick4 + " " + strconv.Itoa(r4))

	L.Players[v.Nick1].FinishGame(gameId, r1, true)
	L.Players[v.Nick2].FinishGame(gameId, r2, true)
	L.Players[v.Nick3].FinishGame(gameId, r3, true)
	L.Players[v.Nick4].FinishGame(gameId, r4, true)
	return "removed"
}

func (L *Ladder) GameContested(nickname string, gameId int) string {
	v, found := L.Games[gameId]
	if !found {
		return "no such game"
	}
	f, _ := v.wasPlayer(nickname)
	if !f {
		return "you're not player"
	}
	if !v.Finished {
		return "finish first"
	}
	if v.Contested {
		return "already contested by " + v.WhoContested
	}
	L.Games[gameId].Contested = true
	L.Games[gameId].WhoContested = nickname
	return "contested"
}

func (L *Ladder) ForceGameContested(nickname string, gameId int) string {
	v, found := L.Games[gameId]
	if !found {
		return "no such game"
	}
	if !v.Finished {
		return "finish first"
	}
	if v.Contested {
		return "already contested by " + v.WhoContested
	}
	L.Games[gameId].Contested = true
	L.Games[gameId].WhoContested = nickname
	return "contested"
}

func (L *Ladder) ShowBlacklisted() string {
	var res []string
	for nickname, player := range L.Players {
		if !player.CanPlay() {
			res = append(res, nickname)
		}
	}
	return strings.Join(res, ",")
}

func (L *Ladder) MyUnfinishedGames(nickname string) string {
	return dump_slice(-1, L.Players[nickname].OpenedGames)
}

func (L *Ladder) AddYellowCard(nickname string) {
	_, found := L.Players[nickname]
	if found {
		L.Players[nickname].AddYellowCard()
	}
}

func (L *Ladder) RemoveYellowCard(nickname string) {
	_, found := L.Players[nickname]
	if found {
		L.Players[nickname].RemoveYellowCard()
	}
}

func (L *Ladder) GetGameInfo(gameId int) string{
	_, found := L.Games[gameId]
	if found {
		return L.Games[gameId].GameInfo()
	}
	return "No such game"
}

func (L *Ladder) GetPlayerInfo(nickname string) string{
	_, found := L.Players[nickname]
	if found {
		return L.Players[nickname].Info()
	}
	return "No such player"
}

func (L *Ladder) GetLeaderBoard() string {
	var checked []IntStringPair
	for nick, player := range L.Players {
		checked = append(checked, IntStringPair{kk:player.Raiting, vv:nick})
	}

	kek := func(p1, p2 *IntStringPair) bool {
		return p1.kk < p2.kk || (p1.kk == p2.kk && p1.vv <= p2.vv)
	}
	By(kek).Sort(checked)
	res := "Leadersboard:"
	t := 0
	i := len(checked)
	for  i > 0 && t < 6 {
		i -= 1
		t += 1
		res += "\n" + checked[i].vv + " " + strconv.Itoa(checked[i].kk)
	}
	return res
}

func (L *Ladder) CanPlay(nickname string) bool{
	_, found := L.Players[nickname]
	if found {
		return L.Players[nickname].CanPlay()
	}
	return true
}
