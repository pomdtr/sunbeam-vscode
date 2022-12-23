#!/usr/bin/env python3

import json
import pathlib
import sqlite3
import urllib.parse

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
    if "folderUri" not in project:
        continue
    uri = urllib.parse.urlparse(project["folderUri"])
    if uri.scheme != "file":
        continue

    path = pathlib.Path(urllib.parse.unquote(uri.path))
    uri = f"vscode://file{path}"

    print(
        json.dumps(
            {
                "title": path.name,
                "subtitle": str(path),
                "actions": [
                    {
                        "type": "open-url",
                        "title": "Open Project",
                        "url": uri,
                        "silent": True,
                    }
                ],
            }
        )
    )
