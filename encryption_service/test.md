# Encryption Microservice

A high-performance Go microservice for secure encryption and decryption operations using AES-256-GCM. This service implements envelope encryption to protect sensitive user data, ensuring that only authorized users can access their information—even from internal administrators.

## Overview

The Encryption Microservice is designed to enhance the security of sensitive user data by ensuring it remains protected from unauthorized access, including from privileged roles such as database administrators. By implementing strong encryption and access controls, the system guarantees that even if the database is compromised, the stored data remains unreadable and unusable to unauthorized individuals.

### Key Features

- **Envelope Encryption**: Implements a two-layer encryption approach with Data Encryption Keys (DEK) and Key Encryption Keys (KEK)
- **AES-256-GCM Encryption/Decryption**: Uses industry-standard encryption algorithms
- **Role-Based Access Control**: Configures different access levels for shared and user-specific data
- **Zero-Knowledge Design**: Prevents administrators from accessing decrypted data
- **Secure Key Management**: Integrates with external Key Management Systems (KMS)
- **Password-Based Access**: Uses user passwords to derive encryption keys
- **Session-Based Memory Handling**: Clears sensitive keys from memory after use

## Project Structure

```
/
├── cmd/                    # Application entry points
│   └── server/             # Main server application
├── common/                 # Shared code across features
│   ├── domain/            # Common domain entities
│   ├── interfaces/        # Common interfaces
│   ├── pkg/              # Common packages
│   └── utils/            # Utility functions
├── internal/              # Private application code
│   ├── di/               # Dependency injection
│   ├── encryption/       # Encryption feature
│   │   ├── domain/      # Domain layer
│   │   │   ├── interfaces/ # Domain interfaces
│   │   │   └── models/  # Domain models
│   │   ├── dtos/        # Data transfer objects
│   │   ├── infrastructure/ # External systems integration
│   │   ├── interfaces/  # Interface adapters
│   │   │   ├── handlers/ # HTTP handlers
│   │   │   └── usecases/ # Use cases
│   │   └── mappers/     # Object mappers
│   ├── framework/        # Framework configurations
│   └── mock/            # Mock implementations
├── pkg/                 # Public libraries
│   ├── crypto/          # Cryptographic utilities
│   ├── kms/             # Key Management System client
│   ├── logger/          # Logging utilities
│   └── errors/          # Error handling
└── scripts/             # Utility scripts
    └── ai_commit_message.py  # AI-powered commit message generator
```

## API Endpoints

### Health Check
- `GET /health` - Check service health

### Encryption Operations
- `POST /encrypt` - Encrypt data using the appropriate DEK
  - Request: `{ "data": "string", "contextId": "string" }`
  - Response: `{ "encryptedData": "string", "edek": "string" }`

- `POST /decrypt` - Decrypt data using the user's KEK
  - Request: `{ "encryptedData": "string", "edek": "string", "contextId": "string" }`
  - Response: `{ "data": "string" }`

- `POST /generate-edek` - Generate Encrypted DEK for new users or data
  - Request: `{ "contextId": "string" }`
  - Response: `{ "edek": "string" }`

## Implementation Details

### Encryption Flow

1. **DEK Generation**: The service generates a Data Encryption Key (DEK) for encrypting user data
2. **KEK Generation**: A Key Encryption Key (KEK) is derived from the user's password
3. **EDEK Creation**: The DEK is encrypted with the KEK to create an Encrypted DEK (EDEK)
4. **Data Encryption**: User data is encrypted using the DEK
5. **Storage**: The encrypted data and EDEK are stored separately

### Decryption Flow

1. **KEK Retrieval**: The KEK is derived from the user's password
2. **DEK Recovery**: The EDEK is decrypted using the KEK to recover the DEK
3. **Data Decryption**: The encrypted data is decrypted using the recovered DEK
4. **Memory Cleanup**: All keys are cleared from memory after use

### Key Management

The service integrates with external Key Management Systems (KMS) to securely store and manage keys:

- **KEK Storage**: Key Encryption Keys are stored in the KMS with role-based access control
- **Role-Based Access**: Different roles (admin, org, user) have different access levels to keys
- **Key Rotation**: Supports key rotation and password changes without data loss

## Configuration

The service can be configured using environment variables:

```env
# Server Configuration
PORT=8080
ENV=development

# Logging
LOG_LEVEL=debug

# KMS Configuration
KMS_ENDPOINT=http://localhost:9000
KMS_REGION=us-east-1
KMS_PROVIDER=hashicorp  # Options: aws, gcp, vault, hashicorp

# Security
ENCRYPTION_KEY_SIZE=32  # 256 bits for AES-256
MAX_REQUEST_SIZE=10485760  # 10MB
PASSWORD_RESET_DAYS=90  # Password reset interval
```

## Getting Started

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Set up environment variables:
   ```bash
   cp .env.example .env
   ```
4. Run the service:
   ```bash
   go run cmd/server/main.go
   ```

## Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o encryption-service cmd/server/main.go
```

## Security Considerations

- **Zero-Knowledge Design**: Even database administrators cannot access decrypted data
- **Envelope Encryption**: Uses a two-layer encryption approach for enhanced security
- **Key Separation**: Encryption keys are stored separately from encrypted data
- **Memory Protection**: Sensitive keys are cleared from memory after use
- **Input Validation**: All inputs are validated and sanitized
- **Rate Limiting**: Prevents brute force attacks
- **HTTPS Enforcement**: All communications are encrypted in transit

## AI Commit Message Generator

This project includes an AI-powered commit message generator that helps create consistent, descriptive commit messages following the Conventional Commits specification.

### Features

- Automatically analyzes git diff to understand changes
- Extracts scope from branch name (e.g., `feature/api` → scope: `api`)
- Generates commit messages in the format: `<type>(<scope>): <description>`
- Provides options to accept, reject, or edit the generated message
- Supports additional context for more accurate message generation

### Usage

1. Stage your changes:
   ```bash
   git add .
   ```

2. Run the AI commit message generator:
   ```bash
   python scripts/ai_commit_message.py
   ```

3. Follow the interactive prompts to:
   - Accept the generated message
   - Reject and exit
   - Generate another message
   - Add additional context
   - Edit the message manually

### Configuration

The script requires a Gemini API key, which should be set as an environment variable:

```bash
export GEMINI_API_KEY='your-api-key'
```

Or add it to your `.env` file:

```
GEMINI_API_KEY=your-api-key
```

### Branch Naming Conventions

The script extracts scope from branch names using these patterns:
- `feature/scope-name` (e.g., `feature/api`)
- `fix/scope-name` (e.g., `fix/auth`)
- `scope/feature-name` (e.g., `api/authentication`)
- `scope-name/feature` (e.g., `api-service/update`)

## License

MIT License

## Getting started

To make it easy for you to get started with GitLab, here's a list of recommended next steps.

Already a pro? Just edit this README.md and make it your own. Want to make it easy? [Use the template at the bottom](#editing-this-readme)!

## Add your files

- [ ] [Create](https://docs.gitlab.com/ee/user/project/repository/web_editor.html#create-a-file) or [upload](https://docs.gitlab.com/ee/user/project/repository/web_editor.html#upload-a-file) files
- [ ] [Add files using the command line](https://docs.gitlab.com/ee/gitlab-basics/add-file.html#add-a-file-using-the-command-line) or push an existing Git repository with the following command:

```
cd existing_repo
git remote add origin https://gitlab.thewitslab.com/shubhpreet.rana/encryption_microservice.git
git branch -M main
git push -uf origin main
```

## Integrate with your tools

- [ ] [Set up project integrations](https://gitlab.thewitslab.com/shubhpreet.rana/encryption_microservice/-/settings/integrations)

## Collaborate with your team

- [ ] [Invite team members and collaborators](https://docs.gitlab.com/ee/user/project/members/)
- [ ] [Create a new merge request](https://docs.gitlab.com/ee/user/project/merge_requests/creating_merge_requests.html)
- [ ] [Automatically close issues from merge requests](https://docs.gitlab.com/ee/user/project/issues/managing_issues.html#closing-issues-automatically)
- [ ] [Enable merge request approvals](https://docs.gitlab.com/ee/user/project/merge_requests/approvals/)
- [ ] [Set auto-merge](https://docs.gitlab.com/ee/user/project/merge_requests/merge_when_pipeline_succeeds.html)

## Test and Deploy

Use the built-in continuous integration in GitLab.

- [ ] [Get started with GitLab CI/CD](https://docs.gitlab.com/ee/ci/quick_start/index.html)
- [ ] [Analyze your code for known vulnerabilities with Static Application Security Testing (SAST)](https://docs.gitlab.com/ee/user/application_security/sast/)
- [ ] [Deploy to Kubernetes, Amazon EC2, or Amazon ECS using Auto Deploy](https://docs.gitlab.com/ee/topics/autodevops/requirements.html)
- [ ] [Use pull-based deployments for improved Kubernetes management](https://docs.gitlab.com/ee/user/clusters/agent/)
- [ ] [Set up protected environments](https://docs.gitlab.com/ee/ci/environments/protected_environments.html)

***

# Editing this README

When you're ready to make this README your own, just edit this file and use the handy template below (or feel free to structure it however you want - this is just a starting point!). Thanks to [makeareadme.com](https://www.makeareadme.com/) for this template.

## Suggestions for a good README

Every project is different, so consider which of these sections apply to yours. The sections used in the template are suggestions for most open source projects. Also keep in mind that while a README can be too long and detailed, too long is better than too short. If you think your README is too long, consider utilizing another form of documentation rather than cutting out information.

## Name
Choose a self-explaining name for your project.

## Description
Let people know what your project can do specifically. Provide context and add a link to any reference visitors might be unfamiliar with. A list of Features or a Background subsection can also be added here. If there are alternatives to your project, this is a good place to list differentiating factors.

## Badges
On some READMEs, you may see small images that convey metadata, such as whether or not all the tests are passing for the project. You can use Shields to add some to your README. Many services also have instructions for adding a badge.

## Visuals
Depending on what you are making, it can be a good idea to include screenshots or even a video (you'll frequently see GIFs rather than actual videos). Tools like ttygif can help, but check out Asciinema for a more sophisticated method.

## Installation
Within a particular ecosystem, there may be a common way of installing things, such as using Yarn, NuGet, or Homebrew. However, consider the possibility that whoever is reading your README is a novice and would like more guidance. Listing specific steps helps remove ambiguity and gets people to using your project as quickly as possible. If it only runs in a specific context like a particular programming language version or operating system or has dependencies that have to be installed manually, also add a Requirements subsection.

## Usage
Use examples liberally, and show the expected output if you can. It's helpful to have inline the smallest example of usage that you can demonstrate, while providing links to more sophisticated examples if they are too long to reasonably include in the README.

## Support
Tell people where they can go to for help. It can be any combination of an issue tracker, a chat room, an email address, etc.

## Roadmap
If you have ideas for releases in the future, it is a good idea to list them in the README.

## Contributing
State if you are open to contributions and what your requirements are for accepting them.

For people who want to make changes to your project, it's helpful to have some documentation on how to get started. Perhaps there is a script that they should run or some environment variables that they need to set. Make these steps explicit. These instructions could also be useful to your future self.

You can also document commands to lint the code or run tests. These steps help to ensure high code quality and reduce the likelihood that the changes inadvertently break something. Having instructions for running tests is especially helpful if it requires external setup, such as starting a Selenium server for testing in a browser.

## Authors and acknowledgment
Show your appreciation to those who have contributed to the project.

## License
For open source projects, say how it is licensed.

## Project status
If you have run out of energy or time for your project, put a note at the top of the README saying that development has slowed down or stopped completely. Someone may choose to fork your project or volunteer to step in as a maintainer or owner, allowing your project to keep going. You can also make an explicit request for maintainers.
