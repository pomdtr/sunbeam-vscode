package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"

	sunbeam "github.com/pomdtr/sunbeam/types"
)

const QUERY = "SELECT json_extract(value, '$.entries') as entries FROM ItemTable WHERE key = 'history.recentlyOpenedPathsList'"

type Project struct {
	FileUri         string `json:"fileUri"`
	FolderUri       string `json:"folderUri"`
	Label           string `json:"label"`
	RemoteAuthority string `json:"remoteAuthority"`
}

func getDatabasePath() string {
	homeDir := os.Getenv("HOME")
	return path.Join(homeDir, "Library", "Application Support", "Code", "User", "globalStorage", "state.vscdb")
}

func main() {
	dbPath := getDatabasePath()
	db, err := exec.Command("sqlite3", dbPath, QUERY).Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var recents []Project
	json.Unmarshal(db, &recents)

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

		cleanPath := strings.Replace(folderUri.Path, os.Getenv("HOME"), "~", 1)

		item := sunbeam.ListItem{
			Title: path.Base(folderUri.Path),
			Accessories: []string{
				cleanPath,
			},
			Actions: []sunbeam.Action{
				{
					Title:  "Open",
					Type:   sunbeam.OpenAction,
					Target: entryUri.String(),
				},
			},
		}

		items = append(items, item)
	}

	json.NewEncoder(os.Stdout).Encode(sunbeam.Page{
		Type:  sunbeam.ListPage,
		Items: items,
	})
}
