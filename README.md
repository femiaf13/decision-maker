# Aproval Voting project

## Commands

- Tailwind `./tailwindcss -i static/tailwind.css -o static/style.css --watch`
- Generate static pages `templ generate`
- Run web server `go run .`

## Building a container
  ```
  **IMPORTANT: You need to build with tailwind and templ before building the container image**
  ```
- `podman image build --tag <whatever you want> .`
- Have a compose file/quadlet file, cmdline args, etc ready to give the app access to a sqlite db via a volume
