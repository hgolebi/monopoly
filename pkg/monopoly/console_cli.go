package monopoly

import (
	"fmt"
	"log"

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
		if response.PropertyId < 0 {
			fmt.Println("You quit.")
			response.Action = QUIT
		}
		return
	}
}

func chooseProperty(properties []int) int {
	for {
		fmt.Println("Choose property (index):")
		page := 0
		max_page := (len(properties) - 1) / 8
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
				fmt.Println("Canceled.")
				return -1
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
