package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
)

func updateVersion(repoPath string, serviceName string) error {
	versionPath := fmt.Sprintf("%s/version.txt", repoPath)
	versionFile := "version.txt"

	newVersion := getVersion(serviceName)

	if repoPath == "/data/alaskartv/androidtv-ci" {
		versionPath = fmt.Sprintf("%s/release.txt", repoPath)
		versionFile = "release.txt"
	}

	os.WriteFile(versionPath, []byte(newVersion), 0644)
	cmds := [][]string{
		{"git", "-C", repoPath, "add", versionFile},
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
