## WS Hub

Generic apps to host and manage ws connections

- Show connected clients (ok)
- Show Number of total “rooms” (ok)
- Show Number of total Messages (ok)
- Show last N Messages (ok)
- Store Messages in database (ok)
- Secret on headers to Connect (Optional)
- Deploy to fly

-- How to run build?
- `go run -o app ./main.go`

-- How to run styles?
- `bunx tailwindcss -i pkg/views/input.css -o public/output.css --watch`
