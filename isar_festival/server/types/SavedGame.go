package types

import "strconv"

type SavedGame struct {
	GameID int
	Nick1 string
	Rait1 int
	Nick2 string
	Rait2 int
	Nick3 string
	Rait3 int
	Nick4 string
	Rait4 int
	Finished bool
	Contested bool
	WhoReported string
	WhoContested string
	Revoked bool
	winner_side int
}

func newGame(gameid int, n1,n2,n3,n4 string, r1,r2,r3,r4 int) *SavedGame {
	p := SavedGame{
		GameID: gameid,
		Nick1: n1,
		Nick2: n2,
		Nick3: n3,
		Nick4: n4,
		Rait1: r1,
		Rait2: r2,
		Rait3: r3,
		Rait4: r4,
		Finished: false,
		Contested: false,
		Revoked: false,
		winner_side: -1,
		WhoReported: "",
		WhoContested: "",
	}
	return &p
}

func (sg *SavedGame) wasPlayer(nickname string) (bool, int) {
	if nickname == sg.Nick1 || nickname == sg.Nick4 {
		return true, 1
	}
	if nickname == sg.Nick3 || nickname == sg.Nick2 {
		return true, 2
	}
	return false, -1
}
//func loadGame(dumpedGame string) *SavedGame {
//	// TODO
//	p := SavedGame{Nickname: name, Raiting: 1500, YellowCardCount:0}
//	return &p
//}

func (sg *SavedGame) FlagDump() string {
	var s1, s2, s3 string
	if sg.Revoked {
		s1 = "T"
	} else {
		s1 = "F"
	}
	if sg.Contested {
		s1 = "T"
	} else {
		s1 = "F"
	}
	if sg.Finished {
		s1 = "T"
	} else {
		s1 = "F"
	}
	return s1 + " " + s2 + " " + s3
}

func (sg *SavedGame) Dump() string {
	if sg.Revoked {
		return "host4 " + sg.Nick1 + " " + sg.Nick2 + " " + sg.Nick3 + " " + sg.Nick4 + " " + sg.WhoReported + " " + sg.WhoContested + "\n"
	} else if sg.Contested {
		return "host3 " + sg.Nick1 + " " + sg.Nick2 + " " + sg.Nick3 + " " + sg.Nick4 + " " + sg.WhoReported + " " + sg.WhoContested + "\n"
	} else if sg.Finished {
		return "host2 " + sg.Nick1 + " " + sg.Nick2 + " " + sg.Nick3 + " " + sg.Nick4 + " " + sg.WhoReported + "\n"
	}
	return "host " + sg.Nick1 + " " + sg.Nick2 + " " + sg.Nick3 + " " + sg.Nick4 + "\n"
}

func (sg *SavedGame) GameInfo() string {
	if sg.Revoked {
		return "Game is removed from games history"
	}
	res := "Game id = " + strconv.Itoa(sg.GameID)
	if sg.Finished {
		if sg.Contested {
			res += " Contested by " + sg.WhoContested
		} else {
			res += " Finished by " + sg.WhoReported
		}
	} else {
		res += " Not Finished "
	}
	res += " played by " + sg.Nick1 + " " +  sg.Nick2 + " " + sg.Nick3 + " " + sg.Nick4 +"\n"
	return res
}

func (sg *SavedGame) FinishGame(side int) (int, int, int, int) {
	sg.Finished = true
	sg.winner_side = side
	return sg.CountWin(side)
}

func (sg *SavedGame) CountWin(side int) (int, int, int, int){
	if side == 1 {
		var pie int
		pie = (50 * (sg.Rait2 + sg.Rait3))/ (sg.Rait1 + sg.Rait2 +  sg.Rait3 + sg.Rait4)
		var p1Win, p2Win, p3Win, p4Win int
		p1Win = pie * sg.Rait4 / (sg.Rait1 + sg.Rait4)
		p4Win = pie * sg.Rait1 / (sg.Rait1 + sg.Rait4)
		p2Win = -1 * pie * sg.Rait3 / (sg.Rait2 + sg.Rait3)
		p3Win = -1 * pie * sg.Rait2 / (sg.Rait2 + sg.Rait3)
		return p1Win, p2Win, p3Win, p4Win
	} else {
		var pie int
		pie = (50 * (sg.Rait1 + sg.Rait4))/ (sg.Rait1 + sg.Rait2 +  sg.Rait3 + sg.Rait4)
		var p1Win, p2Win, p3Win, p4Win int
		p1Win = -1 * pie * sg.Rait4 / (sg.Rait1 + sg.Rait4)
		p4Win = -1 * pie * sg.Rait1 / (sg.Rait1 + sg.Rait4)
		p2Win = pie * sg.Rait3 / (sg.Rait2 + sg.Rait3)
		p3Win = pie * sg.Rait2 / (sg.Rait2 + sg.Rait3)
		return p1Win, p2Win, p3Win, p4Win
	}
}
