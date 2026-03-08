package cli

import "fmt"

func (c *Cli) HandleRoleCommand(commandParams []string) {
	// At this point we know the first command was "role"

	switch commandParams[1] {
	case "make":
		if len(commandParams) < 3 {
			fmt.Println("Usage: make role <role_name>")
			return
		}
		roleName := commandParams[2]
		role, err := c.RoleService.CreateRole(roleName)
		if err != nil {
			fmt.Printf("Error creating role: %v\n", err)

		}
		fmt.Printf("Created role: %v\n", role)
	case "delete":
		if len(commandParams) < 3 {
			fmt.Println("Usage: deleterole <role_name>")
			return
		}
		roleName := commandParams[2]
		err := c.RoleService.DeleteRole(roleName)
		if err != nil {
			fmt.Printf("Error deleting role: %v\n", err)
			return
		}
		fmt.Printf("Deleted role: %v\n", roleName)
	case "ls", "list":
		roles, err := c.RoleService.GetAllRoles()
		if err != nil {
			fmt.Printf("Error listing roles: %v\n", err)
			return
		}
		for _, role := range roles {
			fmt.Printf("Role: %v\n", role)
		}
	}
}
