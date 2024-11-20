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
dispatched to the App Wrapper. App Wrapper broadcasts the message it recieved to all active
Bubble Tea programs.
