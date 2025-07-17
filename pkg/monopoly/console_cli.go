package monopoly

import (
	"fmt"
	"log"

	"github.com/eiannone/keyboard"
)

type ConsoleCLI struct{}

func (c *ConsoleCLI) GetStdAction(state GameState, available []StdAction) StdAction {
	if err := keyboard.Open(); err != nil {
		log.Fatal(err)
	}
	defer keyboard.Close()

	fmt.Println("Available actions:")
	for idx, action := range available {
		fmt.Printf("%v. %s\n", idx, stdActionNames[action])
	}

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			log.Fatal(err)
		}
		if key == keyboard.KeyEsc {
			panic("User quit the game")
		}

		if char < '0' || char >= '0'+rune(len(available)) {
			fmt.Println("Unknown action")
			continue
		}
		action := available[char-'0']
		fmt.Printf("Selected action: %s\n", stdActionNames[action])
		return action
	}
}

func chooseProperty(state GameState, available []int) int {
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
