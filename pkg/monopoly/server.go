package monopoly

import (
	"encoding/json"
	"fmt"
	"net"
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
	State          GameState
	StdActionList  FullActionList
	JailActionList []JailAction
	PropertyId     int
	Price          int
}

type PlayerConn struct {
	Conn net.Conn
	Id   int
}

type ConsoleServer struct {
	Players []PlayerConn
}

func (s *ConsoleServer) Init() int {
	ln, err := net.Listen("tcp", ":12345")
	if err != nil {
		panic(err)
	}
	expectedPlayers := 2
	fmt.Printf("Serwer nasłuchuje na :12345, oczekuje %d graczy...\n", expectedPlayers)
	for len(s.Players) < expectedPlayers {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Błąd połączenia:", err)
			continue
		}
		player := PlayerConn{Conn: conn, Id: len(s.Players)}
		s.Players = append(s.Players, player)
		fmt.Printf("Gracz %d dołączył\n", player.Id)

		// Wyślij ID gracza do klienta jako int
		encoder := json.NewEncoder(conn)
		if err := encoder.Encode(player.Id); err != nil {
			fmt.Println("Błąd wysyłania ID gracza")
			panic(err)
		}
	}
	fmt.Println("Wszyscy gracze połączeni!")
	return expectedPlayers
}

func (s *ConsoleServer) GetStdAction(player int, state GameState, availableActions FullActionList) ActionDetails {
	req := ActionRequest{
		Type:          GetStdAction,
		PlayerId:      player,
		State:         state,
		StdActionList: availableActions,
	}
	encoder := json.NewEncoder(s.Players[player].Conn)
	decoder := json.NewDecoder(s.Players[player].Conn)
	if err := encoder.Encode(req); err != nil {
		fmt.Println("Błąd wysyłania żądania do gracza:", err)
		panic(err)
	}

	var resp ActionDetails
	err := decoder.Decode(&resp)
	if err != nil {
		fmt.Println("Błąd dekodowania odpowiedzi:", err)
		panic("Nie można odczytać odpowiedzi od gracza")
	}
	fmt.Printf("Gracz %d wybrał akcję: %s\n", player, StdActionNames[resp.Action])
	return resp
}

func (s *ConsoleServer) GetJailAction(player int, state GameState, available []JailAction) JailAction {
	req := ActionRequest{
		Type:           GetJailAction,
		PlayerId:       player,
		State:          state,
		JailActionList: available,
	}
	encoder := json.NewEncoder(s.Players[player].Conn)
	decoder := json.NewDecoder(s.Players[player].Conn)
	if err := encoder.Encode(req); err != nil {
		fmt.Println("Błąd wysyłania żądania do gracza:", err)
		panic(err)
	}

	var resp JailAction
	err := decoder.Decode(&resp)
	if err != nil {
		fmt.Println("Błąd dekodowania odpowiedzi:", err)
		panic("Nie można odczytać odpowiedzi od gracza")
	}
	fmt.Printf("Gracz %d wybrał akcję w więzieniu: %s\n", player, JailActionNames[resp])
	return resp
}

func (s *ConsoleServer) BuyDecision(player int, state GameState, propertyId int) bool {
	req := ActionRequest{
		Type:       BuyDecision,
		PlayerId:   player,
		State:      state,
		PropertyId: propertyId,
	}

	encoder := json.NewEncoder(s.Players[player].Conn)
	decoder := json.NewDecoder(s.Players[player].Conn)
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

func (s *ConsoleServer) BuyFromPlayerDecision(player int, state GameState, propertyId int, price int) bool {
	req := ActionRequest{
		Type:       BuyFromPlayerDecision,
		PlayerId:   player,
		State:      state,
		PropertyId: propertyId,
		Price:      price,
	}
	encoder := json.NewEncoder(s.Players[player].Conn)
	decoder := json.NewDecoder(s.Players[player].Conn)
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

func (s *ConsoleServer) SellToPlayerDecision(player int, state GameState, propertyId int, price int) bool {
	req := ActionRequest{
		Type:       SellToPlayerDecision,
		PlayerId:   player,
		State:      state,
		PropertyId: propertyId,
		Price:      price,
	}
	encoder := json.NewEncoder(s.Players[player].Conn)
	decoder := json.NewDecoder(s.Players[player].Conn)
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

func (s *ConsoleServer) BiddingDecision(player int, state GameState, propertyId int, currentPrice int, currentWinner int) int {
	req := ActionRequest{
		Type:       BiddingDecision,
		PlayerId:   player,
		State:      state,
		PropertyId: propertyId,
		Price:      currentPrice,
	}
	encoder := json.NewEncoder(s.Players[player].Conn)
	decoder := json.NewDecoder(s.Players[player].Conn)
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

func (s *ConsoleServer) Finish(f FinishOption, winner int, state GameState) {
	switch f {
	case WIN:
		fmt.Printf("Gra skonczona. Gracz o ID %d wygrał!\n", winner)
	case DRAW:
		fmt.Println("Gra zakończona remisem!")
	case ROUND_LIMIT:
		fmt.Printf("Gra zakończona z powodu przekroczenia limitu rund. Wygrywa gracz z ID %d!\n", winner)
	}
	for _, player := range s.Players {
		player.Conn.Close()
	}
}
