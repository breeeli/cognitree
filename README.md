# cognitree

## WSL Development

Run this project from inside WSL instead of PowerShell/UNC paths.

1. Use the repo's Node version in WSL:
   `nvm use`
2. Start PostgreSQL in WSL:
   `docker compose up -d`
3. Install frontend dependencies once:
   `make frontend-install`
4. Start backend and frontend in separate WSL terminals:
   `make backend`
   `make frontend`

The frontend dev server binds to `0.0.0.0`, so you can open the address Vite prints in your browser.
