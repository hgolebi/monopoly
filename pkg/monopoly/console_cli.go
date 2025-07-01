package monopoly

import (
	"fmt"
	"log"
	"slices"

	"github.com/eiannone/keyboard"
)

type ConsoleCLI struct{}

func (c *ConsoleCLI) GetAction(actions FullActionList, state GameState) (response ActionDetails) {
	if err := keyboard.Open(); err != nil {
		log.Fatal(err)
	}
	defer keyboard.Close()

	fmt.Println("Available actions:")
	for idx, action := range actions.Actions {
		fmt.Printf("%v. %s\n", idx+1, actionNames[action])
	}

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			log.Fatal(err)
		}
		if key == keyboard.KeyEsc {
			fmt.Println("You quit.")
			response.Action = QUIT
			return response
		}

		if char < '1' || char >= '1'+rune(len(actions.Actions)) {
			fmt.Println("Unknown action")
			continue
		}
		action := actions.Actions[char-'1']
		fmt.Printf("Selected action: %s\n", actionNames[action])
		response.Action = action
		switch action {
		case MORTGAGE:
			response.PropertyId = chooseProperty(actions.MortgageList)
		case BUYOUT:
			response.PropertyId = chooseProperty(actions.BuyOutList)
		case BUYHOUSE:
			response.PropertyId = chooseProperty(actions.BuyHouseList)
		case SELLHOUSE:
			response.PropertyId = chooseProperty(actions.SellHouseList)
		}
		return
	}
}

func chooseProperty(properties []int) int {
	fmt.Println("Choose property (index):")
	for _, property := range properties {
		fmt.Printf("Property index: %d\n", property)
	}
	var propertyIndex int
	for {
		char, _, err := keyboard.GetKey()
		if err != nil {
			log.Fatal(err)
		}
		propertyIndex = int(char)
		if slices.Contains(properties, propertyIndex) {
			return propertyIndex
		}
		fmt.Println("Invalid property index. Try again.")
	}
}
