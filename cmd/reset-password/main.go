package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

var (
	conffile = flag.String("c", "", "config yaml file")
	username = flag.String("u", "admin", "username to reset password")
	password = flag.String("p", "toughradius", "new password")
)

func main() {
	flag.Parse()

	if *conffile == "" {
		fmt.Println("Usage: reset-password -c <config-file> [-u <username>] [-p <new-password>]")
		fmt.Println("Example: reset-password -c toughradius.yml -u admin -p newpassword")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Load configuration
	_config := config.LoadConfig(*conffile)

	// Initialize the application
	application := app.NewApplication(_config)
	application.Init(_config)
	defer application.Release()

	// Compute the hash for the new password
	hashedPassword := common.Sha256HashWithSalt(*password, common.SecretSalt)

	// Find the user
	var operator domain.SysOpr
	result := application.DB().Where("username = ?", *username).First(&operator)
	if result.Error != nil {
		log.Fatalf("Failed to find user '%s': %v", *username, result.Error)
	}

	// Update the password
	result = application.DB().Model(&operator).Update("password", hashedPassword)
	if result.Error != nil {
		log.Fatalf("Failed to update password: %v", result.Error)
	}

	fmt.Printf("✓ Password reset successfully for user '%s'\n", *username)
	fmt.Printf("  Username: %s\n", *username)
	fmt.Printf("  New Password: %s\n", *password)
	fmt.Printf("  Level: %s\n", operator.Level)
	fmt.Println("\nYou can now login with the new password.")
}
