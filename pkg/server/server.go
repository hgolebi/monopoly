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

// NegotiationResponse is sent by the client for BuyFromPlayerDecision and SellToPlayerDecision.
// Continue=false means immediate withdrawal/rejection; Continue=true means Price is the response.
type NegotiationResponse struct {
	Continue bool
	Price    int
}

type ActionRequest struct {
	Type             RequestType
	PlayerId         int
	State            monopoly.GameState
	StdActionList    monopoly.FullActionList
	JailActionList   []monopoly.JailAction
	PropertyId       int
	Price            int
	TradingPartnerId int
}

type PlayerIO interface {
	GetStdAction(player int, state monopoly.GameState, availableActions monopoly.FullActionList) monopoly.ActionDetails
	GetJailAction(player int, state monopoly.GameState, available []monopoly.JailAction) monopoly.JailAction
	BuyDecision(player int, state monopoly.GameState, propertyId int) bool
	BuyFromPlayerDecision(player int, state monopoly.GameState, propertyId int, sellerOffer int, tradingPartnerId int) (bool, int)
	SellToPlayerDecision(player int, state monopoly.GameState, propertyId int, buyerOffer int, tradingPartnerId int) (bool, int)
	BiddingDecision(player int, state monopoly.GameState, propertyId int, currentPrice int, currentWinner int) int
}

type PlayerInfo struct {
	isHuman bool
	conn    net.Conn
	bot     PlayerIO
}

type ConsoleServer struct {
	PlayersInfoMap map[int]PlayerInfo
	listener       net.Listener
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

	fmt.Printf("Server listening on :12345, waiting for %d players...\n", humanPlayers)
	for i := range humanPlayers {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}
		id := perm[botPlayers+i]
		playerMap[id] = PlayerInfo{
			isHuman: true,
			conn:    conn,
		}
		fmt.Printf("Player %d joined\n", id)

		encoder := json.NewEncoder(conn)
		if err := encoder.Encode(id); err != nil {
			fmt.Println("Error sending player ID")
			panic(err)
		}
	}
	fmt.Println("All players connected!")
	return &ConsoleServer{
		PlayersInfoMap: playerMap,
		listener:       ln,
	}
}

// reconnect blocks until some client connects, then treats that connection as
// the replacement for the given (disconnected) player. It does not verify the
// identity of the reconnecting client — whoever connects next takes the slot.
func (s *ConsoleServer) reconnect(player int) {
	fmt.Printf("Player %d disconnected. Waiting for a client to reconnect...\n", player)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Println("Connection error while waiting for reconnection:", err)
			continue
		}
		encoder := json.NewEncoder(conn)
		if err := encoder.Encode(player); err != nil {
			fmt.Println("Error sending player ID to reconnected client:", err)
			conn.Close()
			continue
		}
		info := s.PlayersInfoMap[player]
		info.conn = conn
		s.PlayersInfoMap[player] = info
		fmt.Printf("Player %d reconnected\n", player)
		return
	}
}

// sendAndDecode sends req to player's connection and decodes the response into resp.
// If the connection has dropped (send or decode error), it waits for a reconnection
// and retries instead of panicking.
func (s *ConsoleServer) sendAndDecode(player int, req ActionRequest, resp interface{}) {
	for {
		playerInfo := s.PlayersInfoMap[player]
		encoder := json.NewEncoder(playerInfo.conn)
		decoder := json.NewDecoder(playerInfo.conn)
		if err := encoder.Encode(req); err != nil {
			fmt.Printf("Error sending request to player %d: %v\n", player, err)
			playerInfo.conn.Close()
			s.reconnect(player)
			continue
		}
		if err := decoder.Decode(resp); err != nil {
			fmt.Printf("Error decoding response from player %d: %v\n", player, err)
			playerInfo.conn.Close()
			s.reconnect(player)
			continue
		}
		return
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
	var resp monopoly.ActionDetails
	s.sendAndDecode(player, req, &resp)
	fmt.Printf("Player %d chose action: %s\n", player, monopoly.StdActionNames[resp.Action])
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
	var resp monopoly.JailAction
	s.sendAndDecode(player, req, &resp)
	fmt.Printf("Player %d chose jail action: %s\n", player, monopoly.JailActionNames[resp])
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
	var resp bool
	s.sendAndDecode(player, req, &resp)
	fmt.Printf("Player %d decided to buy: %t\n", player, resp)
	return resp
}

func (s *ConsoleServer) BuyFromPlayerDecision(player int, state monopoly.GameState, propertyId int, sellerOffer int, tradingPartnerId int) (bool, int) {
	playerInfo := s.PlayersInfoMap[player]
	if !playerInfo.isHuman {
		return playerInfo.bot.BuyFromPlayerDecision(player, state, propertyId, sellerOffer, tradingPartnerId)
	}
	req := ActionRequest{
		Type:             BuyFromPlayerDecision,
		PlayerId:         player,
		State:            state,
		PropertyId:       propertyId,
		Price:            sellerOffer,
		TradingPartnerId: tradingPartnerId,
	}
	var resp NegotiationResponse
	s.sendAndDecode(player, req, &resp)
	fmt.Printf("Player %d responded to counteroffer: continue=%v price=%d$\n", player, resp.Continue, resp.Price)
	return resp.Continue, resp.Price
}

func (s *ConsoleServer) SellToPlayerDecision(player int, state monopoly.GameState, propertyId int, buyerOffer int, tradingPartnerId int) (bool, int) {
	playerInfo := s.PlayersInfoMap[player]
	if !playerInfo.isHuman {
		return playerInfo.bot.SellToPlayerDecision(player, state, propertyId, buyerOffer, tradingPartnerId)
	}
	req := ActionRequest{
		Type:             SellToPlayerDecision,
		PlayerId:         player,
		State:            state,
		PropertyId:       propertyId,
		Price:            buyerOffer,
		TradingPartnerId: tradingPartnerId,
	}
	var resp NegotiationResponse
	s.sendAndDecode(player, req, &resp)
	fmt.Printf("Player %d responded to buy offer: continue=%v price=%d$\n", player, resp.Continue, resp.Price)
	return resp.Continue, resp.Price
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
	var resp int
	s.sendAndDecode(player, req, &resp)
	fmt.Printf("Player %d made a bid: %d\n", player, resp)
	return resp
}

func (s *ConsoleServer) Finish(f monopoly.FinishOption, winner int, state monopoly.GameState) {
	switch f {
	case monopoly.WIN:
		fmt.Printf("Game over. Player with ID %d won!\n", winner)
	case monopoly.DRAW:
		fmt.Println("Game ended in a draw!")
	case monopoly.ROUND_LIMIT:
		fmt.Printf("Game ended due to round limit. Player with ID %d wins!\n", winner)
	}
	for _, playerInfo := range s.PlayersInfoMap {
		if playerInfo.isHuman {
			playerInfo.conn.Close()
		}
	}
}
