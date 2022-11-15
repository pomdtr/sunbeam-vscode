#!/usr/bin/env python3

import json
import pathlib
import sqlite3

db = (
    pathlib.Path.home()
    / "Library"
    / "Application Support"
    / "Code"
    / "User"
    / "globalStorage"
    / "state.vscdb"
)

# Connect to the database
conn = sqlite3.connect(db)
c = conn.cursor()

# Get the list of projects
c.execute(
    "SELECT json_extract(value, '$.entries') as entries FROM ItemTable WHERE key = 'history.recentlyOpenedPathsList'"
)

res = c.fetchone()
projects = json.loads(res[0])

for project in projects:
    if "label" not in project:
        continue
    if "folderUri" not in project:
        continue
    folderUri: str = project["folderUri"]
    if not folderUri.startswith("file://"):
        continue

    title = folderUri.split("/")[-1]

    print(
        json.dumps(
            {
                "title": title,
                "subtitle": project["label"],
                "actions": [
                    {
                        "type": "open-url",
                        "title": "Open Project",
                        "application": "Visual Studio Code",
                        "url": project["folderUri"],
                    }
                ],
            }
        )
    )
