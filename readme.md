## WS Hub

Generic apps to host and manage ws connections

- Show connected clients (ok)
- Show Number of total “rooms” (ok)
- Show Number of total Messages (ok)
- Show last N Messages (ok)
- Store Messages in database (ok)
- Secret on headers to Connect (Optional)
- Deploy to fly

---

### Details

- Data is stored into a .sqlite3 local file
   - You can delete the file and it will be recreated when the app runs
- It's possible to add authentication for both user accessing it and the ws endpoint

---

- How to run build?
  - `go run -o app ./main.go`

- How to run styles?
  - `bunx tailwindcss -i pkg/views/input.css -o public/output.css --watch`
