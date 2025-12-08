package consoleCLI

import (
	"encoding/json"
	"fmt"
	"log"
	"monopoly/pkg/monopoly"
	"monopoly/pkg/server"
	"net"

	"github.com/eiannone/keyboard"
)

type ConsoleCLI struct {
	ID int
}

func (c *ConsoleCLI) GetStdAction(player int, state monopoly.GameState, availableActions monopoly.FullActionList) monopoly.ActionDetails {
	if err := keyboard.Open(); err != nil {
		log.Fatal(err)
	}
	defer keyboard.Close()

	fmt.Println("Available actions:")
	for idx, action := range availableActions.Actions {
		fmt.Printf("%v. %s\n", idx, monopoly.StdActionNames[action])
	}

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			log.Fatal(err)
		}
		if key == keyboard.KeyEsc {
			panic("User quit the game")
		}
		if char == 's' || char == 'S' {
			fmt.Println(state)
			continue
		}
		if char < '0' || char >= '0'+rune(len(availableActions.Actions)) {
			fmt.Println("Unknown action")
			continue
		}
		action := availableActions.Actions[char-'0']
		fmt.Printf("Selected action: %s\n", monopoly.StdActionNames[action])
		var response monopoly.ActionDetails
		response.Action = action
		switch action {
		case monopoly.MORTGAGE:
			response.PropertyId = chooseProperty(availableActions.MortgageList, state)
		case monopoly.BUYOUT:
			response.PropertyId = chooseProperty(availableActions.BuyOutList, state)
		case monopoly.SELLOFFER:
			response.PropertyId = chooseProperty(availableActions.SellPropertyList, state)
			response.Players = choosePlayers(state.Players, state.CurrentPlayerIdx, state)
			response.Price = choosePrice()
		case monopoly.BUYOFFER:
			response.PropertyId = chooseProperty(availableActions.BuyPropertyList, state)
			response.Price = choosePrice()
		case monopoly.BUYHOUSE:
			response.PropertyId = chooseProperty(availableActions.BuyHouseList, state)
		case monopoly.SELLHOUSE:
			response.PropertyId = chooseProperty(availableActions.SellHouseList, state)
		}
		return response
	}
}

func chooseProperty(properties []int, state monopoly.GameState) int {
	page := 0
	max_page := (len(properties) - 1) / 8
	for {
		fmt.Println("Choose property (index):")
		for idx, property := range properties[page*8 : min(page*8+8, len(properties))] {
			fmt.Printf("%d. Property index: %d\n", idx+1, property)
		}
		if page > 0 {
			fmt.Println("9. Previous page")
		}
		if page < max_page {
			fmt.Println("0. Next page")
		}
		for {
			char, key, err := keyboard.GetKey()
			if err != nil {
				log.Fatal(err)
			}
			if key == keyboard.KeyEsc {
				panic("User quit the game")
			}
			if char == 's' || char == 'S' {
				fmt.Println(state)
				continue
			}
			if page > 0 && char == '9' {
				page--
				break
			}
			if page < max_page && char == '0' {
				page++
				break
			}
			chosen_number := int(char - '1')
			if page*8+chosen_number < len(properties) {
				return properties[page*8+chosen_number]
			}
			fmt.Println("Invalid character. Try again.")
		}
	}
}

func choosePlayers(players []*monopoly.Player, currPlayerIdx int, state monopoly.GameState) []int {
	for {
		var availablePlayers []int
		for idx, player := range players {
			if !player.IsBankrupt && idx != currPlayerIdx {
				availablePlayers = append(availablePlayers, idx)
			}
		}
		fmt.Println("Select players:")
		for idx, player_id := range availablePlayers {
			fmt.Printf("%d. %s\n", idx+1, players[player_id].Name)
		}
		fmt.Println("Press Enter to confirm selection")
		chosenPlayersMap := map[int]bool{}
		for _, player_id := range availablePlayers {
			chosenPlayersMap[player_id] = false
		}
		fmt.Println("Chosen players: []")
		for {
			char, key, err := keyboard.GetKey()
			if err != nil {
				log.Fatal(err)
			}
			if key == keyboard.KeyEsc {
				panic("User quit the game")
			}

			if char == 's' || char == 'S' {
				fmt.Println(state)
				continue
			}

			chosen_number := int(char - '1')
			if chosen_number >= 0 && chosen_number < len(availablePlayers) {
				currDecision := chosenPlayersMap[availablePlayers[chosen_number]]
				chosenPlayersMap[availablePlayers[chosen_number]] = !currDecision
				if !currDecision {
					fmt.Printf("Player %s added\n", players[availablePlayers[chosen_number]].Name)
				} else {
					fmt.Printf("Player %s removed\n", players[availablePlayers[chosen_number]].Name)
				}
			}
			var chosenPlayers []int
			for player_id, selected := range chosenPlayersMap {
				if selected {
					chosenPlayers = append(chosenPlayers, player_id)
				}
			}
			fmt.Println("Chosen players: ", chosenPlayers)
			if key == keyboard.KeyEnter {
				return chosenPlayers
			}
		}
	}
}

func choosePrice() int {
	for {
		fmt.Println("Enter price:")
		var price int
		_, err := fmt.Scanf("%d", &price)
		if err != nil || price < 0 {
			fmt.Println("Invalid price. Try again.")
			continue
		}
		return price
	}
}

func (c *ConsoleCLI) GetJailAction(player int, state monopoly.GameState, available []monopoly.JailAction) monopoly.JailAction {
	if err := keyboard.Open(); err != nil {
		log.Fatal(err)
	}
	defer keyboard.Close()

	fmt.Println("Available jail actions:")
	for idx, action := range available {
		fmt.Printf("%v. %s\n", idx, monopoly.JailActionNames[action])
	}

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			log.Fatal(err)
		}
		if key == keyboard.KeyEsc {
			panic("User quit the game")
		}

		if char == 's' || char == 'S' {
			fmt.Println(state)
			continue
		}
		if char < '0' || char >= '0'+rune(len(available)) {
			fmt.Println("Unknown action")
			continue
		}
		action := available[char-'0']
		fmt.Printf("Selected action: %s\n", monopoly.JailActionNames[action])
		return action
	}
}

func (c *ConsoleCLI) BuyDecision(player int, state monopoly.GameState, propertyId int) bool {
	if err := keyboard.Open(); err != nil {
		log.Fatal(err)
	}
	defer keyboard.Close()

	fmt.Printf("Player %d, do you want to buy property %d? (y/n) \n", player, propertyId)
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			log.Fatal(err)
		}
		if key == keyboard.KeyEsc {
			panic("User quit the game")
		}
		switch char {
		case 's', 'S':
			fmt.Println(state)
		case 'y', 'Y':
			return true
		case 'n', 'N':
			return false
		default:
			fmt.Println("Invalid input. Please enter 'y' or 'n'.")
		}
	}
}

func (c *ConsoleCLI) BuyFromPlayerDecision(player int, state monopoly.GameState, propertyId int, price int) bool {
	if err := keyboard.Open(); err != nil {
		log.Fatal(err)
	}
	defer keyboard.Close()

	fmt.Printf("Player %s wants to sell you property %d for %d? (y/n) \n", state.Players[state.CurrentPlayerIdx].Name, propertyId, price)
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			log.Fatal(err)
		}
		if key == keyboard.KeyEsc {
			panic("User quit the game")
		}
		switch char {
		case 's', 'S':
			fmt.Println(state)
		case 'y', 'Y':
			return true
		case 'n', 'N':
			return false
		default:
			fmt.Println("Invalid input. Please enter 'y' or 'n'.")
		}
	}
}

func (c *ConsoleCLI) SellToPlayerDecision(player int, state monopoly.GameState, propertyId int, price int) bool {
	if err := keyboard.Open(); err != nil {
		log.Fatal(err)
	}
	defer keyboard.Close()

	fmt.Printf("Player %s wants to buy property %d from you for %d. Do you want to sell it? (y/n)\n ", state.Players[state.CurrentPlayerIdx].Name, propertyId, price)
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			log.Fatal(err)
		}
		if key == keyboard.KeyEsc {
			panic("User quit the game")
		}
		switch char {
		case 's', 'S':
			fmt.Println(state)
		case 'y', 'Y':
			return true
		case 'n', 'N':
			return false
		default:
			fmt.Println("Invalid input. Please enter 'y' or 'n'.")
		}
	}
}

func (c *ConsoleCLI) BiddingDecision(player int, state monopoly.GameState, propertyId int, currentPrice int) int {
	if err := keyboard.Open(); err != nil {
		log.Fatal(err)
	}
	defer keyboard.Close()

	fmt.Printf("Auction bidding for property %d. Current price is %d. Do you want to bid? (y/n)\n", propertyId, currentPrice)
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			log.Fatal(err)
		}
		if key == keyboard.KeyEsc {
			panic("User quit the game")
		}
		switch char {
		case 's', 'S':
			fmt.Println(state)
		case 'y', 'Y':
			bid := choosePrice()
			return bid
		case 'n', 'N':
			return 0
		default:
			fmt.Println("Invalid input. Please enter 'y' or 'n'.")
		}
	}
}

func StartClient() {
	c := &ConsoleCLI{}
	conn, err := net.Dial("tcp", "localhost:12345")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)
	if err := decoder.Decode(&c.ID); err != nil {
		panic(err)
	}
	fmt.Printf("Connected to server with ID: %d\n", c.ID)
	fmt.Println("Press 's' to show current game state at any time.")
	for {
		var req server.ActionRequest
		if err := decoder.Decode(&req); err != nil {
			fmt.Println("Failed to decode request")
			panic(err)
		}
		var resp interface{}
		switch req.Type {
		case server.GetStdAction:
			resp = c.GetStdAction(req.PlayerId, req.State, req.StdActionList)
		case server.GetJailAction:
			resp = c.GetJailAction(req.PlayerId, req.State, req.JailActionList)
		case server.BuyDecision:
			resp = c.BuyDecision(req.PlayerId, req.State, req.PropertyId)
		case server.BuyFromPlayerDecision:
			resp = c.BuyFromPlayerDecision(req.PlayerId, req.State, req.PropertyId, req.Price)
		case server.SellToPlayerDecision:
			resp = c.SellToPlayerDecision(req.PlayerId, req.State, req.PropertyId, req.Price)
		case server.BiddingDecision:
			resp = c.BiddingDecision(req.PlayerId, req.State, req.PropertyId, req.Price)

		default:
			panic(fmt.Sprintf("Unknown request type: %v", req.Type))
		}
		encoder.Encode(resp)
	}
}
