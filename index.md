---
layout: default
---

# Welcome to EnvTrack CLI

EnvTrack CLI is a command-line interface tool for managing and tracking development server usage within a team. It allows team members to interact with the EnvTrack service, managing organizations, projects, environments, and variables.

## Key Features

- **Authentication**: Secure access to the EnvTrack service
- **Organization Management**: List and manage organizations
- **Project Management**: Handle projects within organizations
- **Environment Tracking**: Keep track of environments within projects
- **Variable Management**: Manage variables within environments
- **Flexible Output**: Support for JSON, YAML, and CSV output formats
- **Secure Configuration**: Local storage of authentication tokens and settings

## Quick Start

### Installation

Choose your preferred installation method:

#### Using Go Install

```bash
go install github.com/envtrack/envtrack-cli/cmd/envtrack@latest
```

#### Homebrew (macOS and Linux)

```bash
brew tap envtrack/tap
brew install envtrack
```

#### Scoop (Windows)

```powershell
scoop bucket add yourbucket https://github.com/envtrack/scoop-bucket.git
scoop install envtrack
```

#### Docker

```bash
docker pull envtrack/envtrack:latest
docker run envtrack/envtrack:latest [commands]
```

### Basic Usage

1. Authenticate with the EnvTrack service:
   ```bash
   envtrack auth <your-auth-token>
   ```

2. List organizations:
   ```bash
   envtrack organizations
   ```

3. List projects for an organization:
   ```bash
   envtrack projects --organizationId <org-id>
   ```

4. List environments for a project:
   ```bash
   envtrack environments --organizationId <org-id> --projectId <project-id>
   ```

5. List variables for an environment:
   ```bash
   envtrack variables --organizationId <org-id> --projectId <project-id> --environmentId <env-id>
   ```

## Documentation

For more detailed information, please check out our [full documentation](https://github.com/envtrack/envtrack-cli/wiki).

## Contributing

We welcome contributions! Please see our [contribution guidelines](https://github.com/envtrack/envtrack-cli/blob/main/CONTRIBUTING.md) for more details.

## Support

If you encounter any issues or have questions, please [open an issue](https://github.com/envtrack/envtrack-cli/issues) on our GitHub repository.

## License

EnvTrack CLI is released under the MIT License. See the [LICENSE](https://github.com/envtrack/envtrack-cli/blob/main/LICENSE) file for more details.