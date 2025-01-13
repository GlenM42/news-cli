# news-cli

This is a project with the goal of learning some Golang. The idea consists of an ssh application that will
have two parts to it:

1. A section for browsing news
2. A section for multi-user chat

So far, two parts have been implemented separately.

## Architecture of the app

- The central component is the **Wish SSH Server**. It handles all incoming SSH connections.
- Then we have an **App Wrapper** that acts as a manager for all active client sessions.
  - It maintains a list of active Bubble Tea programs, one per connected user.
  - Provides functionality to broadcast messages to all connected sessions.
- **Bubble Teas Program** is the lowest unit of the whole Architecture.
  - It represents a unique terminal UI session.
  - Handles user-specific input, such as typing messages or switching tabs.
  - Updates its own view based on the user actions.

Therefore, the user message, once sent in the Bubble Tea program, is
dispatched to the App Wrapper. App Wrapper broadcasts the message it received to all active
Bubble Tea programs.

## Documentation for SSH Chat Application

The application uses the following libraries:
- **Bubble Tea**: To create the TUI.
- **Wish**: To handle SSH sessions.
- **Lipgloss**: For styling the TUI.
- **Termenv**: For terminal manipulation.

### **File Structure**

```
.
├── main.go                // Entry point for the application
├── server.go              // Contains the SSH server configuration and startup logic
├── auth.go                // Handles user authentication logic
├── model.go               // Contains TUI state, update logic, and rendering
├── styles.go              // Styling definitions for the TUI
├── authorized_keys.json   // External file storing user credentials (username and public keys)
├── .env                   // Stores environment variables for keyboard-interactive authentication
├── go.mod                 // Go module file
└── go.sum                 // Dependency checksum file
```

### **Installation Instructions**

1. **Pre-requisites**:
    - Go (1.18+)
    - SSH client (to connect to the server)
    - `authorized_keys.json` (see below for format)

2. **Setting up Environment Variables**:
   Create a `.env` file in the root directory:
   ```
   NEWS_API_KEY=...
   SSH_HOST=...
   SSH_PORT=...
   QUESTION_1=...
   QUESTION_2=...
   ```

3. **Adding Authorized Users**:
   Create a JSON file named `authorized_keys.json`:
   ```json
   {
     "users": [
       {
         "username": "...",
         "publicKey": "..."
       },
       {
         "username": "...",
         "publicKey": "..."
       }
     ]
   }
   ```

4. **Running the Application**:
    - Fetch dependencies and build the application:
      ```bash
      go mod tidy
      go build
      ```
    - Run the application:
      ```bash
      ./your-app-binary
      ```

5. **Connecting via SSH**:
   Use an SSH client to connect:
   ```bash
   ssh -p 23234 <username>@localhost
   ```