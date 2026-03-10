package cli

import (
	"context"
	"fmt"
)

func (c *Cli) HandleUserRoleCOmmands(commandParams []string) {
	// At this point we know the first command was "ur"

	switch commandParams[1] {
	case "assign":
		if len(commandParams) < 4 {
			fmt.Println("Usage: assignuserrole <username> <role_name>")
			return
		}
		username := commandParams[2]
		roleName := commandParams[3]
		err := c.UserService.AssignUserToRole(context.Background(), username, roleName)
		if err != nil {
			fmt.Printf("Error assigning user to role: %v\n", err)
			return
		}
		fmt.Printf("Assigned user %v to role %v\n", username, roleName)
	case "remove":
		if len(commandParams) < 4 {
			fmt.Println("Usage: removeuserrole <username> <role_name>")
			return
		}
		username := commandParams[2]
		roleName := commandParams[3]
		err := c.UserService.RemoveUserFromRole(context.Background(), username, roleName)
		if err != nil {
			fmt.Printf("Error removing user from role: %v\n", err)
			return
		}
		fmt.Printf("Removed user %v from role %v\n", username, roleName)
	case "ls", "list":
		if len(commandParams) < 3 {
			fmt.Println("Usage: listuserroles <username>")
			return
		}
		username := commandParams[2]
		roles, err := c.UserService.GetUserRolesByUsername(context.Background(), username)
		if err != nil {
			fmt.Printf("Error listing user roles: %v\n", err)
			return
		}
		fmt.Printf("Roles for user %v:\n", username)
		for _, role := range roles {
			fmt.Printf("- %v\n", role)
		}
	}
}
