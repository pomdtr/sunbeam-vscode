package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"

	sunbeam "github.com/pomdtr/sunbeam/types"
	_ "modernc.org/sqlite"
)

const QUERY = "SELECT json_extract(value, '$.entries') as entries FROM ItemTable WHERE key = 'history.recentlyOpenedPathsList'"

type Project struct {
	FileUri         string `json:"fileUri"`
	FolderUri       string `json:"folderUri"`
	Label           string `json:"label"`
	RemoteAuthority string `json:"remoteAuthority"`
}

func getDatabasePath(homeDir string) (string, bool) {
	switch runtime.GOOS {
	case "darwin":
		return path.Join(homeDir, "Library", "Application Support", "Code", "User", "globalStorage", "state.vscdb"), true
	case "linux":
		return path.Join(homeDir, ".config", "Code", "User", "globalStorage", "state.vscdb"), true
	case "windows":
		return path.Join(homeDir, "AppData", "Roaming", "Code", "User", "globalStorage", "state.vscdb"), true
	default:
		return "", false
	}
}

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	dbPath, ok := getDatabasePath(homeDir)
	if !ok {
		fmt.Fprintln(os.Stderr, "Unsupported OS")
		os.Exit(1)
	}

	conn, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer conn.Close()

	row := conn.QueryRow(QUERY)

	var data []byte
	if err = row.Scan(&data); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var recents []Project
	if err := json.Unmarshal(data, &recents); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	items := make([]sunbeam.ListItem, 0)
	for _, recent := range recents {
		if recent.FolderUri == "" {
			continue
		}
		folderUri, err := url.Parse(recent.FolderUri)
		if err != nil {
			continue
		}

		entryUri := url.URL{
			Scheme: "vscode",
			Host:   "file",
			Path:   folderUri.Path,
		}

		cleanPath := strings.Replace(folderUri.Path, homeDir, "~", 1)

		item := sunbeam.ListItem{
			Title: path.Base(folderUri.Path),
			Accessories: []string{
				cleanPath,
			},
			Actions: []sunbeam.Action{
				sunbeam.NewOpenAction("Open", entryUri.String()),
			},
		}

		items = append(items, item)
	}

	json.NewEncoder(os.Stdout).Encode(sunbeam.Page{
		Title: "Recent Projects",
		Type:  sunbeam.ListPage,
		Items: items,
	})
}
