package server

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"monopoly/pkg/monopoly"
	"net"
	"time"
)

type RequestType int

const (
	GetStdAction RequestType = iota
	GetJailAction
	BuyDecision
	BuyFromPlayerDecision
	SellToPlayerDecision
	BiddingDecision
)

type ActionRequest struct {
	Type           RequestType
	PlayerId       int
	State          monopoly.GameState
	StdActionList  monopoly.FullActionList
	JailActionList []monopoly.JailAction
	PropertyId     int
	Price          int
}

type PlayerIO interface {
	GetStdAction(player int, state monopoly.GameState, availableActions monopoly.FullActionList) monopoly.ActionDetails
	GetJailAction(player int, state monopoly.GameState, available []monopoly.JailAction) monopoly.JailAction
	BuyDecision(player int, state monopoly.GameState, propertyId int) bool
	BuyFromPlayerDecision(player int, state monopoly.GameState, propertyId int, price int) bool
	SellToPlayerDecision(player int, state monopoly.GameState, propertyId int, price int) bool
	BiddingDecision(player int, state monopoly.GameState, propertyId int, currentPrice int, currentWinner int) int
}

type PlayerInfo struct {
	isHuman bool
	conn    net.Conn
	bot     PlayerIO
}

type ConsoleServer struct {
	PlayersInfoMap map[int]PlayerInfo
}

func NewConsoleServer(humanPlayers int, bots []PlayerIO) *ConsoleServer {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	botPlayers := len(bots)
	totalPlayers := humanPlayers + botPlayers
	perm := r.Perm(totalPlayers)
	botIndexes := perm[:botPlayers]
	playerMap := make(map[int]PlayerInfo)
	for i := range botPlayers {
		playerMap[botIndexes[i]] = PlayerInfo{
			isHuman: false,
			bot:     bots[i],
		}
	}

	ln, err := net.Listen("tcp", ":12345")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Serwer nasłuchuje na :12345, oczekuje %d graczy...\n", humanPlayers)
	for i := range humanPlayers {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Błąd połączenia:", err)
			continue
		}
		id := perm[botPlayers+i]
		playerMap[id] = PlayerInfo{
			isHuman: true,
			conn:    conn,
		}
		fmt.Printf("Gracz %d dołączył\n", id)

		// Wyślij ID gracza do klienta jako int
		encoder := json.NewEncoder(conn)
		if err := encoder.Encode(id); err != nil {
			fmt.Println("Błąd wysyłania ID gracza")
			panic(err)
		}
	}
	fmt.Println("Wszyscy gracze połączeni!")
	return &ConsoleServer{
		PlayersInfoMap: playerMap,
	}
}

func (s *ConsoleServer) Init() []string {
	player_names := make([]string, len(s.PlayersInfoMap))
	for id, info := range s.PlayersInfoMap {
		if info.isHuman {
			player_names[id] = fmt.Sprintf("Player_%d", id)
		} else {
			player_names[id] = fmt.Sprintf("Bot_%d", id)
		}
	}
	fmt.Println(player_names)
	return player_names
}

func (s *ConsoleServer) GetStdAction(player int, state monopoly.GameState, availableActions monopoly.FullActionList) monopoly.ActionDetails {
	playerInfo := s.PlayersInfoMap[player]
	if !playerInfo.isHuman {
		return playerInfo.bot.GetStdAction(player, state, availableActions)
	}
	req := ActionRequest{
		Type:          GetStdAction,
		PlayerId:      player,
		State:         state,
		StdActionList: availableActions,
	}
	encoder := json.NewEncoder(playerInfo.conn)
	decoder := json.NewDecoder(playerInfo.conn)
	if err := encoder.Encode(req); err != nil {
		fmt.Println("Błąd wysyłania żądania do gracza:", err)
		panic(err)
	}

	var resp monopoly.ActionDetails
	err := decoder.Decode(&resp)
	if err != nil {
		fmt.Println("Błąd dekodowania odpowiedzi:", err)
		panic("Nie można odczytać odpowiedzi od gracza")
	}
	fmt.Printf("Gracz %d wybrał akcję: %s\n", player, monopoly.StdActionNames[resp.Action])
	return resp
}

func (s *ConsoleServer) GetJailAction(player int, state monopoly.GameState, available []monopoly.JailAction) monopoly.JailAction {
	playerInfo := s.PlayersInfoMap[player]
	if !playerInfo.isHuman {
		return playerInfo.bot.GetJailAction(player, state, available)
	}
	req := ActionRequest{
		Type:           GetJailAction,
		PlayerId:       player,
		State:          state,
		JailActionList: available,
	}
	encoder := json.NewEncoder(playerInfo.conn)
	decoder := json.NewDecoder(playerInfo.conn)
	if err := encoder.Encode(req); err != nil {
		fmt.Println("Błąd wysyłania żądania do gracza:", err)
		panic(err)
	}

	var resp monopoly.JailAction
	err := decoder.Decode(&resp)
	if err != nil {
		fmt.Println("Błąd dekodowania odpowiedzi:", err)
		panic("Nie można odczytać odpowiedzi od gracza")
	}
	fmt.Printf("Gracz %d wybrał akcję w więzieniu: %s\n", player, monopoly.JailActionNames[resp])
	return resp
}

func (s *ConsoleServer) BuyDecision(player int, state monopoly.GameState, propertyId int) bool {
	playerInfo := s.PlayersInfoMap[player]
	if !playerInfo.isHuman {
		return playerInfo.bot.BuyDecision(player, state, propertyId)
	}
	req := ActionRequest{
		Type:       BuyDecision,
		PlayerId:   player,
		State:      state,
		PropertyId: propertyId,
	}

	encoder := json.NewEncoder(playerInfo.conn)
	decoder := json.NewDecoder(playerInfo.conn)
	if err := encoder.Encode(req); err != nil {
		fmt.Println("Błąd wysyłania żądania do gracza:", err)
		panic(err)
	}

	var resp bool
	err := decoder.Decode(&resp)
	if err != nil {
		fmt.Println("Błąd dekodowania odpowiedzi:", err)
		panic("Nie można odczytać odpowiedzi od gracza")
	}
	fmt.Printf("Gracz %d podjął decyzję o zakupie: %t\n", player, resp)
	return resp
}

func (s *ConsoleServer) BuyFromPlayerDecision(player int, state monopoly.GameState, propertyId int, price int) bool {
	playerInfo := s.PlayersInfoMap[player]
	if !playerInfo.isHuman {
		return playerInfo.bot.BuyFromPlayerDecision(player, state, propertyId, price)
	}
	req := ActionRequest{
		Type:       BuyFromPlayerDecision,
		PlayerId:   player,
		State:      state,
		PropertyId: propertyId,
		Price:      price,
	}
	encoder := json.NewEncoder(playerInfo.conn)
	decoder := json.NewDecoder(playerInfo.conn)
	if err := encoder.Encode(req); err != nil {
		fmt.Println("Błąd wysyłania żądania do gracza:", err)
		panic(err)
	}

	var resp bool
	err := decoder.Decode(&resp)
	if err != nil {
		fmt.Println("Błąd dekodowania odpowiedzi:", err)
		panic("Nie można odczytać odpowiedzi od gracza")
	}
	fmt.Printf("Gracz %d podjął decyzję o zakupie od innego gracza: %t\n", player, resp)
	return resp
}

func (s *ConsoleServer) SellToPlayerDecision(player int, state monopoly.GameState, propertyId int, price int) bool {
	playerInfo := s.PlayersInfoMap[player]
	if !playerInfo.isHuman {
		return playerInfo.bot.SellToPlayerDecision(player, state, propertyId, price)
	}
	req := ActionRequest{
		Type:       SellToPlayerDecision,
		PlayerId:   player,
		State:      state,
		PropertyId: propertyId,
		Price:      price,
	}
	encoder := json.NewEncoder(playerInfo.conn)
	decoder := json.NewDecoder(playerInfo.conn)
	if err := encoder.Encode(req); err != nil {
		fmt.Println("Błąd wysyłania żądania do gracza:", err)
		panic(err)
	}

	var resp bool
	err := decoder.Decode(&resp)
	if err != nil {
		fmt.Println("Błąd dekodowania odpowiedzi:", err)
		panic("Nie można odczytać odpowiedzi od gracza")
	}
	fmt.Printf("Gracz %d podjął decyzję o sprzedaży do innego gracza: %t\n", player, resp)
	return resp
}

func (s *ConsoleServer) BiddingDecision(player int, state monopoly.GameState, propertyId int, currentPrice int, currentWinner int) int {
	playerInfo := s.PlayersInfoMap[player]
	if !playerInfo.isHuman {
		return playerInfo.bot.BiddingDecision(player, state, propertyId, currentPrice, currentWinner)
	}
	req := ActionRequest{
		Type:       BiddingDecision,
		PlayerId:   player,
		State:      state,
		PropertyId: propertyId,
		Price:      currentPrice,
	}
	encoder := json.NewEncoder(playerInfo.conn)
	decoder := json.NewDecoder(playerInfo.conn)
	if err := encoder.Encode(req); err != nil {
		fmt.Println("Błąd wysyłania żądania do gracza:", err)
		panic(err)
	}

	var resp int
	err := decoder.Decode(&resp)
	if err != nil {
		fmt.Println("Błąd dekodowania odpowiedzi:", err)
		panic("Nie można odczytać odpowiedzi od gracza")
	}
	fmt.Printf("Gracz %d złożył ofertę licytacyjną: %d\n", player, resp)
	return resp
}

func (s *ConsoleServer) Finish(f monopoly.FinishOption, winner int, state monopoly.GameState) {
	switch f {
	case monopoly.WIN:
		fmt.Printf("Gra skonczona. Gracz o ID %d wygrał!\n", winner)
	case monopoly.DRAW:
		fmt.Println("Gra zakończona remisem!")
	case monopoly.ROUND_LIMIT:
		fmt.Printf("Gra zakończona z powodu przekroczenia limitu rund. Wygrywa gracz z ID %d!\n", winner)
	}
	for _, playerInfo := range s.PlayersInfoMap {
		if playerInfo.isHuman {
			playerInfo.conn.Close()
		}
	}
}
