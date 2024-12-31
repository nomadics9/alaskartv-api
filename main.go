package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func updateVersion(repoPath string, serviceName string) (string, error) {
	versionPath := fmt.Sprintf("%s/version.txt", repoPath)
	newVersion := getVersion(serviceName)

	godotenv.Load(".botenv")
	botToken := os.Getenv("BOT_TOKEN")
	chatid := os.Getenv("CHAT_ID")

	if serviceName == "alaskartv" {
		versionPath = fmt.Sprintf("%s/release.txt", repoPath)
		bumpVersionTv(repoPath)
	}

	currentVersion, err := os.ReadFile(versionPath)
	if err != nil {
		return "", fmt.Errorf("failed to read version file: %w", err)
	}

	// Compare versions
	if strings.TrimSpace(string(currentVersion)) == strings.TrimSpace(newVersion) {
		message := map[string]interface{}{
			"status":  "unchanged",
			"message": fmt.Sprintf("Version is already at %s", newVersion),
			"service": serviceName,
			"version": newVersion,
		}

		jsonResponse, err := json.Marshal(message)
		if err != nil {
			return "", fmt.Errorf("failed to marshal JSON: %w", err)
		}
		return string(jsonResponse), nil
	}

	os.WriteFile(versionPath, []byte(newVersion), 0644)
	cmds := [][]string{
		{"git", "-C", repoPath, "add", "."},
		{"git", "-C", repoPath, "commit", "-m", fmt.Sprintf("Bump version to %s", newVersion)},
		{"git", "-C", repoPath, "push"},
		{
			"curl", "-X", "POST",
			fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken),
			"-d", fmt.Sprintf("chat_id=%s", chatid),
			"-d", fmt.Sprintf("text=<b>Alaskar-api</b>: <b>%s</b> updated to <b>%s</b>", serviceName, newVersion),
			"-d", fmt.Sprintf("parse_mode=HTML"),
		},
	}

	for _, cmd := range cmds {
		err := exec.Command(cmd[0], cmd[1:]...).Run()
		if err != nil {
			return "", fmt.Errorf("failed to execute command %v: %w", cmd, err)
		}
	}

	message := map[string]interface{}{
		"status":  "updated",
		"message": fmt.Sprintf("Successfully updated from %s to %s", string(currentVersion), newVersion),
		"service": serviceName,
		"version": newVersion,
	}

	jsonResponse, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(jsonResponse), nil
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
			newVersion, _ := BumpWithRollover(currentPatch, "patch")
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
	if err := os.WriteFile(filePath, []byte(updatedContent), 0644); err != nil {
		fmt.Errorf("failed to write to %s: %w", filePath, err)
	}

	fmt.Println("Version bumped successfully!")

}

func notifyHandler(c *gin.Context) {
	var data struct {
		Name    string `json:"name"`
		Message string `json:"message"`
	}

	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if data.Name == "" {
		data.Name = "unknown"
	}
	if data.Message == "" {
		data.Message = "unknown"
	}

	godotenv.Load(".botenv")
	botToken := os.Getenv("BOT_TOKEN")
	chatid := os.Getenv("CHAT_ID")

	message := fmt.Sprintf("<b>Alaskar-api</b>: <b>%s</b> \n \n %s", data.Name, data.Message)

	telegramAPI := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	resp, err := http.PostForm(telegramAPI, url.Values{
		"chat_id":    {chatid},
		"text":       {message},
		"parse_mode": {"HTML"},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send Telegram message"})
		return
	}
	defer resp.Body.Close()

	c.JSON(http.StatusOK, gin.H{"message": "Notification sent"})
}

func notify(jsonResponse string) {
	var data struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Service string `json:"service"`
		Version string `json:"version"`
	}
	godotenv.Load(".botenv")
	botToken := os.Getenv("BOT_TOKEN")
	chatid := os.Getenv("CHAT_ID")

	json.Unmarshal([]byte(jsonResponse), &data)

	message := fmt.Sprintf("<b>Alaskar-api</b>: <b>%s %s</b> \n \n %s", data.Service, data.Status, data.Message)
	telegramAPI := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	resp, _ := http.PostForm(telegramAPI, url.Values{
		"chat_id":    {chatid},
		"text":       {message},
		"parse_mode": {"HTML"},
	})
	defer resp.Body.Close()
}

func main() {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	router.POST("/api/alaskarfin", func(c *gin.Context) {
		jsonResponse, err := updateVersion("/data/alaskartv/docker-ci/alaskarfin", "alaskarfin")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, jsonResponse)
		notify(jsonResponse)
	})

	router.POST("/api/alaskarseer", func(c *gin.Context) {
		jsonResponse, err := updateVersion("/data/alaskartv/docker-ci/alaskarseer", "alaskarseer")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, jsonResponse)
		notify(jsonResponse)
	})

	router.POST("/api/alaskartv", func(c *gin.Context) {
		jsonResponse, err := updateVersion("/data/alaskartv/androidtv-ci", "alaskartv")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, jsonResponse)
		notify(jsonResponse)

	})

	router.POST("/notify", notifyHandler)
	router.GET("/", func(c *gin.Context) {
		c.File("index.html")
	})

	fmt.Println("Webhook server running on port 8080")
	router.Run(":8080")
}
