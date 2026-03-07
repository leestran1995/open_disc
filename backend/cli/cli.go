package cli

import (
	"backend/auth"
	"backend/role"
	"backend/room"
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
	RoomService *room.RoomService
}

func NewCli(otc *auth.Otc, roleService *role.Service, userService *user.UserService, roomService *room.RoomService) *Cli {
	return &Cli{
		Otc:         otc,
		RoleService: roleService,
		UserService: userService,
		RoomService: roomService,
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
		case "role":
			c.HandleRoleCommand(commandParams)
			continue
		case "ur":
			c.HandleUserRoleCOmmands(commandParams)
			continue
		case "assignroomrole":
			if len(commandParams) < 3 {
				fmt.Println("Usage: assignroomrole <room_name> <role_name>")
				continue
			}
			roomName := commandParams[1]
			roleName := commandParams[2]
			err := c.RoomService.AssignRoomRole(context.Background(), roomName, roleName)
			if err != nil {
				fmt.Printf("Error assigning room role: %v\n", err)
				continue
			}
			fmt.Printf("Assigned role %v to room %v\n", roleName, roomName)
		case "removeroomrole":
			if len(commandParams) < 3 {
				fmt.Println("Usage: removeroomrole <room_name> <role_name>")
				continue
			}
			roomName := commandParams[1]
			roleName := commandParams[2]
			err := c.RoomService.RemoveRoomRole(context.Background(), roomName, roleName)
			if err != nil {
				fmt.Printf("Error removing room role: %v\n", err)
				continue
			}
			fmt.Printf("Removed role %v from room %v\n", roleName, roomName)
		default:
			fmt.Println("Unknown command")
			fmt.Println("Available commands: otc")
			fmt.Println("---------------")
			fmt.Print("-> ")
			continue
		}

	}
}
