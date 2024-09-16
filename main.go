package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/sys/unix"

	"github.com/inhies/go-bytesize"
)

func getAvailableSpace(dir string) (int64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(dir, &stat); err != nil {
		return 0, err
	}
	return int64(stat.Bavail) * int64(stat.Bsize), nil
}

func sendDiscordNotification(webhookURL, message string) error {
	payloadBytes, err := json.Marshal(map[string]string{"content": message})
	if err != nil {
		return fmt.Errorf("error marshalling JSON payload: %v", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("received non-204 response from Discord: %s - %s", resp.Status, string(bodyBytes))
	}

	return nil
}

func getNotificationTimestampFilePath(webhookURL string) string {
	hash := sha256.Sum256([]byte(webhookURL))
	hashStr := hex.EncodeToString(hash[:])
	return filepath.Join(os.TempDir(), "disk_space_checker_last_notification_"+hashStr)
}

func shouldSendNotification(webhookURL string, cooldown time.Duration) (bool, error) {
	filePath := getNotificationTimestampFilePath(webhookURL)
	//fmt.Println("Timestamp file path:", getNotificationTimestampFilePath(webhookURL))

	// Open the file (or create if it doesn't exist) with read/write permissions
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return false, fmt.Errorf("error opening timestamp file: %v", err)
	}
	defer file.Close()

	// Acquire an exclusive lock
	if err := unix.Flock(int(file.Fd()), unix.LOCK_EX); err != nil {
		return false, fmt.Errorf("error acquiring file lock: %v", err)
	}
	defer unix.Flock(int(file.Fd()), unix.LOCK_UN) // Ensure the lock is released

	// Read the last notification timestamp
	data, err := io.ReadAll(file)
	if err != nil {
		return false, fmt.Errorf("error reading timestamp file: %v", err)
	}

	lastSentStr := string(bytes.TrimSpace(data))
	if lastSentStr == "" {
		// No timestamp recorded yet; we can send the notification
		return true, nil
	}

	lastSentUnix, err := strconv.ParseInt(lastSentStr, 10, 64)
	if err != nil {
		// Could not parse timestamp; assume we can send notification
		return true, nil
	}

	lastSentTime := time.Unix(lastSentUnix, 0)
	if time.Since(lastSentTime) >= cooldown {
		return true, nil
	}

	return false, nil
}

func updateNotificationTimestamp(webhookURL string) error {
	filePath := getNotificationTimestampFilePath(webhookURL)

	// Open the file with write permissions
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening timestamp file: %v", err)
	}
	defer file.Close()

	// Acquire an exclusive lock
	if err := unix.Flock(int(file.Fd()), unix.LOCK_EX); err != nil {
		return fmt.Errorf("error acquiring file lock: %v", err)
	}
	defer unix.Flock(int(file.Fd()), unix.LOCK_UN) // Ensure the lock is released

	// Write the current timestamp
	currentTime := strconv.FormatInt(time.Now().Unix(), 10)
	if _, err := file.WriteString(currentTime); err != nil {
		return fmt.Errorf("error writing timestamp file: %v", err)
	}

	return nil
}

func main() {
	limitFlag := flag.String("limit", "", "Minimum required free space (e.g., 50GB)")
	discordFlag := flag.String("discord", "", "Discord webhook URL for notifications (optional)")
	cooldownFlag := flag.Duration("cooldown", time.Minute, "Cooldown duration between notifications (e.g., 1m, 30s)")
	flag.Parse()

	if *limitFlag == "" {
		fmt.Println("Error: --limit flag is required.")
		flag.Usage()
		os.Exit(2)
	}

	if flag.NArg() < 1 {
		fmt.Println("Error: Directory path is required.")
		flag.Usage()
		os.Exit(2)
	}
	dir := flag.Arg(0)

	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Printf("Error resolving directory path: %v\n", err)
		os.Exit(2)
	}

	stat, err := os.Stat(absDir)
	if err != nil {
		fmt.Printf("Error accessing directory %s: %v\n", absDir, err)
		os.Exit(2)
	}
	if !stat.IsDir() {
		fmt.Printf("Error: Path %s is not a directory.\n", absDir)
		os.Exit(2)
	}

	limitBytes, err := bytesize.Parse(*limitFlag)
	if err != nil {
		fmt.Printf("Error parsing limit size: %v\n", err)
		os.Exit(2)
	}

	availableBytes, err := getAvailableSpace(absDir)
	if err != nil {
		fmt.Printf("Error getting available space: %v\n", err)
		os.Exit(2)
	}

	availableByteSize := bytesize.ByteSize(availableBytes)
	if availableByteSize >= limitBytes {
		fmt.Printf("Sufficient space: %s available.\n", availableByteSize)
		os.Exit(0)
	}

	message := fmt.Sprintf("Warning: Only %s available in %s, which is below the limit of %s.",
		availableByteSize, absDir, limitBytes)
	fmt.Println(message)

	if *discordFlag != "" {
		sendNotification, err := shouldSendNotification(*discordFlag, *cooldownFlag)
		if err != nil {
			fmt.Printf("Error checking notification cooldown: %v\n", err)
		} else if sendNotification {
			if err := sendDiscordNotification(*discordFlag, message); err != nil {
				fmt.Printf("Error sending Discord notification: %v\n", err)
			} else {
				fmt.Println("Discord notification sent successfully.")
				if err := updateNotificationTimestamp(*discordFlag); err != nil {
					fmt.Printf("Error updating notification timestamp: %v\n", err)
				}
			}
		} else {
			fmt.Println("Notification not sent due to rate limiting.")
		}
	}

	os.Exit(1)
}