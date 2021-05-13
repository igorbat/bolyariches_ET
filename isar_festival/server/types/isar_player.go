// literally don't care.
package types

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type IntStringPair struct {
	kk     int
	vv     string
}
// By is the type of a "less" function that defines the ordering of its Planet arguments.
type By func(p1, p2 *IntStringPair) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by By) Sort(planets []IntStringPair) {
	ps := &planetSorter{
		planets: planets,
		by:      by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ps)
}

// planetSorter joins a By function and a slice of Planets to be sorted.
type planetSorter struct {
	planets []IntStringPair
	by      func(p1, p2 *IntStringPair) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *planetSorter) Len() int {
	return len(s.planets)
}

// Swap is part of sort.Interface.
func (s *planetSorter) Swap(i, j int) {
	s.planets[i], s.planets[j] = s.planets[j], s.planets[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *planetSorter) Less(i, j int) bool {
	return s.by(&s.planets[i], &s.planets[j])
}

func find(a []int, id int) (bool, int) {
	flag := false
	idx := 0
	for i, n := range a {
		if id == n {
			flag = true
			idx = i
		}
	}
	return flag, idx
}

func dump_slice(limit int, slicee []int) string {
	var info []int
	if limit != -1 {
		if len(slicee) > 5 {
			info = slicee[len(slicee)-5:]
		} else {
			info = slicee
		}
	} else {
		info = slicee
	}
	var res []string
	for i := range info {
		number := info[i]
		text := strconv.Itoa(number)
		res = append(res, text)
	}
	return strings.Join(res, ", ")
}

type IsarPlayer struct {
	Nickname string
	Raiting int
	YellowCardCount int
	OpenedGames []int
	FinishedGames []int
}

func newIsarPlayer(name string) *IsarPlayer {
	p := IsarPlayer{Nickname: name, Raiting: 1500, YellowCardCount:0}
	return &p
}

//func loadIsarPlayer(dump string) *IsarPlayer {
//	// TODO
//	//p := IsarPlayer{Nickname: name, Raiting: 1500, YellowCardCount:0}
//	//return &p
//}

func (ip *IsarPlayer) Dump() string {
	return ip.Nickname + " " + strconv.Itoa(ip.Raiting) + " " + strconv.Itoa(ip.YellowCardCount) + "\n"
}

func (ip *IsarPlayer) smallGameHistory() string {
	return dump_slice(5, ip.FinishedGames)
}

func (ip *IsarPlayer) StartGame(gameId int) {
	ip.OpenedGames = append(ip.OpenedGames, gameId)
}

func (ip *IsarPlayer) AddYellowCard() {
	ip.YellowCardCount += 1
}

func (ip *IsarPlayer) RemoveYellowCard() {
	ip.YellowCardCount -= 1
}

func (ip *IsarPlayer) FinishGame(gameId int, points int, bad bool) {
	flag, idx := find(ip.OpenedGames, gameId)
	if flag {
		ip.OpenedGames[idx] = ip.OpenedGames[len(ip.OpenedGames) - 1]
		ip.OpenedGames = ip.OpenedGames[:len(ip.OpenedGames) - 1]
		ip.FinishedGames = append(ip.FinishedGames, gameId)
		if !bad {
			ip.Raiting += points
		} else {
			ip.Raiting -= points
		}
	} else {
		if !bad {
			fmt.Println("Something goes wrong with " + ip.Nickname + ": Finishing Game Filed: " + strconv.Itoa(gameId))
		} else {
			ip.Raiting -= points
		}
	}
}

func (ip *IsarPlayer) AddFinishedGame(gameId int) {
	ip.FinishedGames = append(ip.FinishedGames, gameId)
}

func (ip *IsarPlayer) CanPlay() bool {
	return (ip.YellowCardCount < 3) && (len(ip.OpenedGames) < 3)
}
func (ip *IsarPlayer) Info() string {
	return ip.Nickname +
		" Raiting: " + strconv.Itoa(ip.Raiting) +
		" YellowCardsCount:" + strconv.Itoa(ip.YellowCardCount) +
		" FinishedGamesCount:" + strconv.Itoa(len(ip.FinishedGames)) +
		" OpenedGamesCount:" + strconv.Itoa(len(ip.OpenedGames)) + " " + dump_slice(-1, ip.OpenedGames) + "\n"
}


