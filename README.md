# EnvTrack CLI

EnvTrack CLI is a command-line interface tool for managing and tracking development server usage within a team. It allows team members to interact with the EnvTrack service, managing organizations, projects, environments, and variables.

## Features

- Authenticate with the EnvTrack service
- List and manage organizations
- List and manage projects within organizations
- List and manage environments within projects
- List and manage variables within environments
- Multiple output formats (JSON, YAML, CSV) for easy integration with other tools
- Secure local storage of authentication tokens and configuration

## Installation

### Prerequisites

- Go 1.22 or later

### Using Go Install

```bash
go install github.com/envtrack/envtrack-cli/cmd/envtrack@latest
```

### Homebrew (macOS and Linux)

```bash
brew tap envtrack/tap
brew install envtrack
```

### Scoop (Windows)

```powershell
scoop bucket add yourbucket https://github.com/envtrack/scoop-bucket.git
scoop install envtrack
```

### Docker

```bash
docker pull envtrack/envtrack:latest
docker run envtrack/envtrack:latest [commands]
```

## Usage

### Authentication

Before using any commands that interact with the EnvTrack service, you need to authenticate:

```bash
envtrack auth <your-auth-token>
```

### Configuration

Set the default output format:

```bash
envtrack configure set default_format json
```

Supported formats: `json`, `yaml`, `csv`

### Commands

1. List Organizations:
   ```bash
   envtrack organizations
   ```

2. List Projects:
   ```bash
   envtrack projects --organizationId <org-id>
   ```

3. List Environments:
   ```bash
   envtrack environments --organizationId <org-id> --projectId <project-id>
   ```

4. List Variables:
   ```bash
   envtrack variables --organizationId <org-id> --projectId <project-id> --environmentId <env-id>
   ```

### Global Flags

- `--format, -f`: Override the output format for a single command
  ```bash
  envtrack organizations --format yaml
  ```

## Development

### Prerequisites

- Go 1.17 or later
- Make (optional, for using Makefile commands)

### Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/envtrack/envtrack-cli.git
   cd envtrack-cli
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

### Building

To build for your current platform:

```bash
go build -o envtrack ./cmd/envtrack
```

To cross-compile for multiple platforms, use the provided build script:

```bash
./build.sh
```

### Testing

Run the test suite:

```bash
go test ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

If you encounter any problems or have any questions, please open an issue on the GitHub repository.