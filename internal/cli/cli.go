package cli

import (
	"bufio"
	"fmt"
	"open_discord/internal/auth"
	"os"
	"strings"
)

type Cli struct {
	Otc *auth.Otc
}

func (c *Cli) Run() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Waiting for input")
	fmt.Println("---------------")

	for {
		fmt.Print("-> ")
		userText, _ := reader.ReadString('\n')
		text := strings.Replace(userText, "\r\n", "\n", -1)

		switch text {
		case "otc\n":
			otc, err := c.Otc.GenerateUuid()
			if err != nil {
				fmt.Printf("Error generating OTC: %v\n", err)
				continue
			}
			fmt.Println(otc)
		default:
			fmt.Println("Unknown command")
			fmt.Println("Available commands: otc")
			fmt.Println("---------------")
			fmt.Print("-> ")
			continue
		}

	}
}
