# docker-context-size

A command-line tool that displays the contents and sizes of a Docker build context in a tree format. This helps you understand what files are being included in your Docker build context and identify large files that might be slowing down your builds.

## Features

- Displays directory contents in a tree structure with file sizes
- Respects `.dockerignore` rules to show only files that would be included in the Docker context
- Configurable depth limiting to control how deep the tree display goes
- Human-readable file size formatting (B, KB, MB, GB, etc.)

## Installation

### Using Go Install

```bash
go install github.com/artificialinc/docker-context-size@latest
```

## Usage

```bash
# Display current directory with depth 1 (default)
docker-context-size

# Display a specific directory
docker-context-size --directory /path/to/project

# Display with unlimited depth
docker-context-size --depth -1

# Display with custom depth
docker-context-size --directory ./my-project --depth 3

# Short aliases
docker-context-size -d ./my-project -l 2
```

## Command Line Options

- `--directory, -d`: Directory to analyze (default: current directory)
- `--depth, -l`: Maximum depth to display (default: 1, use -1 for unlimited)
- `--help, -h`: Show help message

## Example Output

```bash
docker-context-size (1.2 MB)
├── .git (890.3 KB)
├── pkg (12.8 KB)
│   └── docker (12.8 KB)
├── .dockerignore (45 B)
├── .gitignore (123 B)
├── Dockerfile (1.2 KB)
├── go.mod (456 B)
├── go.sum (12.3 KB)
├── main.go (789 B)
└── README.md (2.1 KB)
```

## How It Works

The tool creates a temporary tar archive of the directory using the same logic Docker uses when building a context, including respecting `.dockerignore` rules. It then analyzes the archive contents to calculate sizes and display them in a tree format.

## License

This project is licensed under the MIT License.
