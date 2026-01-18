# File Share



Self-hosted file sharing service. Files are automatically deleted 24 hours after upload. Designed to run behind Cloudflare using chunked uploads to bypass body size limits.



## Tech Stack

Go (Backend), Vanilla JS + Uppy (Frontend), SQLite/Filesystem.



## Security Note

Ensure the secret key is strong. The application uses CF-Connecting-IP for rate limiting; ensure it runs behind Cloudflare or adjust the middleware in main.go.





## Features:

- TUS Protocol: Resumable, chunked uploads

- Ephemeral Storage: Automatic file deletion after 24h

- Access Control: Token-based authentication for uploads; public access for downloads.

- Security: Rate limiting (IP-based), Security Headers (HSTS, CSP), Non-root Docker user.

