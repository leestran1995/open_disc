package cli

import (
	"backend/auth"
	"backend/role"
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Cli struct {
	Otc         *auth.Otc
	RoleService *role.Service
}

func NewCli(otc *auth.Otc, roleService *role.Service) *Cli {
	return &Cli{
		Otc:         otc,
		RoleService: roleService,
	}
}

func (c *Cli) Run() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Waiting for input")
	fmt.Println("---------------")

	for {
		fmt.Print("-> ")
		userText, _ := reader.ReadString('\n')
		text := strings.Replace(userText, "\r\n", "\n", -1)
		text = strings.ReplaceAll(text, "\n", "")

		commandParams := strings.Split(text, " ")

		switch commandParams[0] {
		case "otc":
			otc, err := c.Otc.GenerateUuid()
			if err != nil {
				fmt.Printf("Error generating OTC: %v\n", err)
				continue
			}
			fmt.Println(otc)
		case "makerole":
			if len(commandParams) < 2 {
				fmt.Println("Usage: makerole <role_name>")
				continue
			}
			roleName := commandParams[1]
			role, err := c.RoleService.CreateRole(roleName)
			if err != nil {
				fmt.Printf("Error creating role: %v\n", err)
				continue
			}
			fmt.Printf("Created role: %v\n", role)
		case "deleterole":
			if len(commandParams) < 2 {
				fmt.Println("Usage: deleterole <role_name>")
				continue
			}
			roleName := commandParams[1]
			err := c.RoleService.DeleteRole(roleName)
			if err != nil {
				fmt.Printf("Error deleting role: %v\n", err)
				continue
			}
			fmt.Printf("Deleted role: %v\n", roleName)
		default:
			fmt.Println("Unknown command")
			fmt.Println("Available commands: otc")
			fmt.Println("---------------")
			fmt.Print("-> ")
			continue
		}

	}
}
