package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gman/internal/config"
	"github.com/fatih/color"
)

// groupCmd represents the group command
var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage repository groups",
	Long: `Manage repository groups for batch operations.
Groups allow you to organize repositories and perform operations on specific sets.

Examples:
  gman group create frontend web-app mobile-app
  gman group list
  gman group add frontend admin-panel
  gman sync --group frontend`,
}

// groupCreateCmd creates a new group
var groupCreateCmd = &cobra.Command{
	Use:   "create <group-name> [repo1] [repo2] ...",
	Short: "Create a new repository group",
	Long: `Create a new repository group with the specified repositories.
All repository aliases must already exist in the configuration.

Examples:
  gman group create frontend web-app mobile-app
  gman group create backend api-server auth-service --desc "Backend services"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runGroupCreate,
}

// groupListCmd lists all groups
var groupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all repository groups",
	Long:  `Display all configured repository groups with their repositories.`,
	Aliases: []string{"ls"},
	RunE:  runGroupList,
}

// groupDeleteCmd deletes a group
var groupDeleteCmd = &cobra.Command{
	Use:     "delete <group-name>",
	Short:   "Delete a repository group",
	Long:    `Delete the specified repository group. This does not affect the repositories themselves.`,
	Aliases: []string{"rm", "remove"},
	Args:    cobra.ExactArgs(1),
	RunE:    runGroupDelete,
}

// groupAddCmd adds repositories to a group
var groupAddCmd = &cobra.Command{
	Use:   "add <group-name> <repo1> [repo2] ...",
	Short: "Add repositories to an existing group",
	Long:  `Add one or more repositories to an existing group.`,
	Args:  cobra.MinimumNArgs(2),
	RunE:  runGroupAdd,
}

// groupRemoveCmd removes repositories from a group
var groupRemoveCmd = &cobra.Command{
	Use:   "remove <group-name> <repo1> [repo2] ...",
	Short: "Remove repositories from a group",
	Long:  `Remove one or more repositories from an existing group.`,
	Args:  cobra.MinimumNArgs(2),
	RunE:  runGroupRemove,
}

var groupDescription string

func init() {
	rootCmd.AddCommand(groupCmd)
	
	// Add subcommands
	groupCmd.AddCommand(groupCreateCmd)
	groupCmd.AddCommand(groupListCmd)
	groupCmd.AddCommand(groupDeleteCmd)
	groupCmd.AddCommand(groupAddCmd)
	groupCmd.AddCommand(groupRemoveCmd)
	
	// Add flags
	groupCreateCmd.Flags().StringVarP(&groupDescription, "desc", "d", "", "Group description")
}

func runGroupCreate(cmd *cobra.Command, args []string) error {
	groupName := args[0]
	repositories := args[1:]

	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create group
	if err := configMgr.CreateGroup(groupName, groupDescription, repositories); err != nil {
		return err
	}

	fmt.Printf("%s Created group '%s' with %d repositories\n", 
		color.GreenString("✅"), groupName, len(repositories))
	
	if groupDescription != "" {
		fmt.Printf("   Description: %s\n", groupDescription)
	}
	
	fmt.Printf("   Repositories: %s\n", strings.Join(repositories, ", "))
	
	return nil
}

func runGroupList(cmd *cobra.Command, args []string) error {
	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	groups := configMgr.GetGroups()
	if len(groups) == 0 {
		fmt.Println("No groups configured. Use 'gman group create' to create groups.")
		return nil
	}

	// Sort groups by name
	var groupNames []string
	for name := range groups {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	fmt.Printf("\n%s (%d groups):\n\n", color.CyanString("Repository Groups"), len(groups))

	for _, name := range groupNames {
		group := groups[name]
		fmt.Printf("%s %s (%d repositories)\n", 
			color.YellowString("📁"), 
			color.GreenString(name), 
			len(group.Repositories))
		
		if group.Description != "" {
			fmt.Printf("   %s\n", color.WhiteString(group.Description))
		}
		
		fmt.Printf("   Repositories: %s\n", 
			color.BlueString(strings.Join(group.Repositories, ", ")))
		
		fmt.Printf("   Created: %s\n\n", 
			group.CreatedAt.Format("2006-01-02 15:04"))
	}

	return nil
}

func runGroupDelete(cmd *cobra.Command, args []string) error {
	groupName := args[0]

	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Delete group
	if err := configMgr.DeleteGroup(groupName); err != nil {
		return err
	}

	fmt.Printf("%s Deleted group '%s'\n", color.GreenString("✅"), groupName)
	return nil
}

func runGroupAdd(cmd *cobra.Command, args []string) error {
	groupName := args[0]
	repositories := args[1:]

	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Add to group
	if err := configMgr.AddToGroup(groupName, repositories); err != nil {
		return err
	}

	fmt.Printf("%s Added %d repositories to group '%s': %s\n", 
		color.GreenString("✅"), len(repositories), groupName, strings.Join(repositories, ", "))
	
	return nil
}

func runGroupRemove(cmd *cobra.Command, args []string) error {
	groupName := args[0]
	repositories := args[1:]

	// Load configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Remove from group
	if err := configMgr.RemoveFromGroup(groupName, repositories); err != nil {
		return err
	}

	fmt.Printf("%s Removed %d repositories from group '%s': %s\n", 
		color.GreenString("✅"), len(repositories), groupName, strings.Join(repositories, ", "))
	
	return nil
}