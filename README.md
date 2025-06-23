## DEPRECATED - I don't know if this works any longer.
# Xerox

Copy Github Commits to Google Spreadsheet.

### Steps
- Create a project in Google Sheets API [console](https://console.developers.google.com/apis/library/sheets.googleapis.com).
- Download `credentials.json` for above project and keep it in project directory.
- Create AccessToken in Github, export it as `GITHUB_ACCESS_TOKEN`
- Update config.yml for configuration regarding repository to copy commits.
- Use `xerox_amd64` for *nix and `darwin` for MacOS.
- Run `go build` to rebuild binary.
