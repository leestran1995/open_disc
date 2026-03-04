package cli

import (
	"backend/auth"
	"backend/role"
	"backend/user"
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

type Cli struct {
	Otc         *auth.Otc
	RoleService *role.Service
	UserService *user.UserService
}

func NewCli(otc *auth.Otc, roleService *role.Service, userService *user.UserService) *Cli {
	return &Cli{
		Otc:         otc,
		RoleService: roleService,
		UserService: userService,
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
		case "assignuserrole":
			if len(commandParams) < 3 {
				fmt.Println("Usage: assignuserrole <username> <role_name>")
				continue
			}
			username := commandParams[1]
			roleName := commandParams[2]
			err := c.UserService.AssignUserToRole(context.Background(), username, roleName)
			if err != nil {
				fmt.Printf("Error assigning user to role: %v\n", err)
				continue
			}
			fmt.Printf("Assigned user %v to role %v\n", username, roleName)
		case "removeuserrole":
			if len(commandParams) < 3 {
				fmt.Println("Usage: removeuserrole <username> <role_name>")
				continue
			}
			username := commandParams[1]
			roleName := commandParams[2]
			err := c.UserService.RemoveUserFromRole(context.Background(), username, roleName)
			if err != nil {
				fmt.Printf("Error removing user from role: %v\n", err)
				continue
			}
			fmt.Printf("Removed user %v from role %v\n", username, roleName)
		default:
			fmt.Println("Unknown command")
			fmt.Println("Available commands: otc")
			fmt.Println("---------------")
			fmt.Print("-> ")
			continue
		}

	}
}
