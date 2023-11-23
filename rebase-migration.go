package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func getMigrationFiles(migrationPath string) []string {
	cmd := exec.Command("git", "log", "--name-only", "--pretty=format:", "--topo-order", migrationPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("git log command failed: %s", err)
	}
	files := strings.Split(string(output), "\n")
	var migrationFiles []string
	for _, file := range files {
		if strings.HasPrefix(file, migrationPath) && (strings.HasSuffix(file, ".up.sql") || strings.HasSuffix(file, ".down.sql")) {
			migrationFiles = append(migrationFiles, file)
		}
	}
	return migrationFiles
}

type MigrationGroup struct {
	Version   int
	BaseNames []string
}

func NewMigrationGroup(version int) MigrationGroup {
	return MigrationGroup{
		Version:   version,
		BaseNames: make([]string, 0),
	}
}

func groupByVersionNumber(migrations []string) []MigrationGroup {
	addedBaseNames := make(map[string]bool)
	migrationGroupsMap := make(map[int]MigrationGroup)
	for _, migration := range migrations {
		versionStr := strings.Split(filepath.Base(migration), "_")[0]
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			log.Fatalf("Invalid version number '%s' in file '%s': %s", versionStr, migration, err)
		}
		_, baseName, _ := strings.Cut(strings.TrimSuffix(strings.TrimSuffix(filepath.Base(migration), ".up.sql"), ".down.sql"), "_")

		if _, ok := addedBaseNames[baseName]; ok {
			continue
		} else {
			addedBaseNames[baseName] = true
		}

		group, exists := migrationGroupsMap[version]
		if !exists {
			group = NewMigrationGroup(version)
		}
		group.BaseNames = append(group.BaseNames, baseName)
		migrationGroupsMap[version] = group
	}

	// Convert map to slice
	migrationGroups := make([]MigrationGroup, 0, len(migrationGroupsMap))
	for _, group := range migrationGroupsMap {
		migrationGroups = append(migrationGroups, group)
	}

	// Sort the slice by version number
	sort.Slice(migrationGroups, func(i, j int) bool {
		return migrationGroups[i].Version < migrationGroups[j].Version
	})

	return migrationGroups
}

func renameMigrations(migrationPath string, migrationGroups []MigrationGroup) {
	highestVersionNumber := migrationGroups[len(migrationGroups)-1].Version
	for _, group := range migrationGroups {
		for i, baseName := range group.BaseNames {
			if i == 0 { // Skip the first file which is from the master branch
				continue
			}
			upOldName := fmt.Sprintf("%s/%06d_%s.up.sql", migrationPath, group.Version, baseName)
			downOldName := fmt.Sprintf("%s/%06d_%s.down.sql", migrationPath, group.Version, baseName)

			if _, err := os.Stat(upOldName); err != nil {
				continue
			}

			highestVersionNumber++
			upNewName := fmt.Sprintf("%s/%06d_%s.up.sql", migrationPath, highestVersionNumber, baseName)
			downNewName := fmt.Sprintf("%s/%06d_%s.down.sql", migrationPath, highestVersionNumber, baseName)

			// Rename .up.sql file
			if err := os.Rename(upOldName, upNewName); err != nil {
				log.Fatalf("Failed to rename %s to %s: %s", upOldName, upNewName, err)
			}
			// Rename .down.sql file
			if err := os.Rename(downOldName, downNewName); err != nil {
				log.Fatalf("Failed to rename %s to %s: %s", downOldName, downNewName, err)
			}

			fmt.Printf("Renamed %s to %s\n", upOldName, upNewName)
			fmt.Printf("Renamed %s to %s\n", downOldName, downNewName)
		}
	}
}

func commitChanges(migrationPath string) {
	commitMessage := "Rename migration files to resolve version number conflicts"

	// Add only the migration files to the staging area.
	cmd := exec.Command("git", "add", migrationPath)
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to stage migration files: %s", err)
	}

	// Check if there are any changes to commit.
	statusCmd := exec.Command("git", "status", "--porcelain", migrationPath)
	statusOutput, err := statusCmd.Output()
	if err != nil {
		log.Fatalf("Failed to check git status: %s", err)
	}

	// If there's output from the status command, there are changes to commit.
	if len(statusOutput) > 0 {
		// Commit the changes.
		commitCmd := exec.Command("git", "commit", "-m", commitMessage)
		if err := commitCmd.Run(); err != nil {
			log.Fatalf("Failed to commit changes: %s", err)
		}
		fmt.Println("Committed changes:", commitMessage)
	} else {
		fmt.Println("No changes to commit for migration files.")
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: rebase-migration <migration_path>")
		os.Exit(1)
	}
	migrationPath := os.Args[1]

	migrations := getMigrationFiles(migrationPath)
	migrationsGroupedByVersionNumbers := groupByVersionNumber(migrations)

	renameMigrations(migrationPath, migrationsGroupedByVersionNumbers)

	commitChanges(migrationPath)
}
