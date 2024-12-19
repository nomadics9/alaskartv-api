package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func updateVersion(repoPath string, serviceName string) error {
	versionPath := fmt.Sprintf("%s/version.txt", repoPath)
	//versionFile := "version.txt"
	newVersion := getVersion(serviceName)

	if serviceName == "alaskartv" {
		versionPath = fmt.Sprintf("%s/release.txt", repoPath)
		//versionFile = "release.txt"
		bumpVersionTv(repoPath)
	}

	os.WriteFile(versionPath, []byte(newVersion), 0644)
	cmds := [][]string{
		{"git", "-C", repoPath, "add", "."},
		{"git", "-C", repoPath, "commit", "-m", fmt.Sprintf("Bump version to %s", newVersion)},
		{"git", "-C", repoPath, "push"},
	}

	for _, cmd := range cmds {
		err := exec.Command(cmd[0], cmd[1:]...).Run()
		if err != nil {
			return fmt.Errorf("failed to execute command %v: %w", cmd, err)
		}
	}

	return nil
}

func bumpVersionTv(repoPath string) {
	filePath := fmt.Sprintf("%s/version.txt", repoPath)
	file, _ := os.Open(filePath)
	defer file.Close()

	var updatedContent string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "VERSION_NAME="):
			parts := strings.Split(line, "=")
			currentPatch := parts[1]
			newVersion, _ := Bump(currentPatch, "patch")
			updatedContent += fmt.Sprintf("VERSION_NAME=%s\n", newVersion)

		case strings.HasPrefix(line, "VERSION_CODE="):
			parts := strings.Split(line, "=")
			currentCode, _ := strconv.Atoi(parts[1])
			newVersionCode := currentCode + 1
			updatedContent += fmt.Sprintf("VERSION_CODE=%d\n", newVersionCode)

		default:
			updatedContent += line + "\n"
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Errorf("error reading %s: %w", filePath, err)
	}
	// Write the updated content back to the file
	if err := os.WriteFile(filePath, []byte(updatedContent), 0644); err != nil {
		fmt.Errorf("failed to write to %s: %w", filePath, err)
	}

	fmt.Println("Version bumped successfully!")

}

func main() {
	router := gin.Default()

	router.POST("/api/alaskarfin", func(c *gin.Context) {
		err := updateVersion("/data/alaskartv/docker-ci/alaskarfin", "alaskarfin")
		if err != nil {
			c.String(http.StatusInternalServerError, "Something went wrong: %s", err.Error())
			return
		}
		c.String(http.StatusOK, "Bumped Alaskarfin!")
	})

	router.POST("/api/alaskarseer", func(c *gin.Context) {
		err := updateVersion("/data/alaskartv/docker-ci/alaskarseer", "alaskarseer")
		if err != nil {
			c.String(http.StatusInternalServerError, "Something went wrong: %s", err.Error())
			return
		}
		c.String(http.StatusOK, "Bumped Alaskarseer!")
	})

	router.POST("/api/alaskartv", func(c *gin.Context) {
		err := updateVersion("/data/alaskartv/androidtv-ci", "alaskartv")
		if err != nil {
			c.String(http.StatusInternalServerError, "Something went wrong: %s", err.Error())
			return
		}
		c.String(http.StatusOK, "Bumped AlaskarTV!")
	})

	fmt.Println("Webhook server running on port 8080")
	router.Run(":8080")
}
