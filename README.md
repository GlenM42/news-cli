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

## Plan for deployment

#### **Step 1: Build the Go Application**

Compile the app into a static binary for Linux:

```bash
GOOS=linux GOARCH=amd64 go build -o multichat
```

We’ll end up with an executable called `multichat`.

#### **Step 2: Create a Dockerfile**

Here’s a Dockerfile for your app:

```dockerfile
# Use a minimal base image for Go apps
FROM golang:1.20-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy the Go modules and source code
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application
RUN go build -o multichat

# Use a lightweight image for running the app
FROM alpine:latest

WORKDIR /root/

# Copy the built binary from the builder
COPY --from=builder /app/multichat .

# Copy the .env file (if required)
COPY .env .

# Expose the SSH port
EXPOSE 23234

# Run the application
CMD ["./multichat"]
```

#### **Step 3: Build and Run the Docker Image**

1. Build the Docker image:

   ```bash
   docker build -t multichat .
   ```

2. Run the container:

   ```bash
   docker run -d -p 23234:23234 --name multichat --env-file .env multichat
   ```

   - `-d`: Runs the container in detached mode.
   - `-p 23234:23234`: Maps port `23234` on the host to the container.
   - `--env-file .env`: Loads the `.env` file into the container.

3. Verify it’s running:
   ```bash
   docker ps
   ```

#### **Step 4: Configure Public Access**

For people across the globe to access the app, we need to:

1. **Open the SSH port (23234) in your firewall**:

   - If using `ufw`:
     ```bash
     sudo ufw allow 23234
     ```
   - For other firewalls, open TCP port 23234.

2. **Point a Domain Name to the Server**:

3. **Use a Public IP (if no domain)**:
   - Share a public IP address:
     ```bash
     ssh -p 23234 user@<your-public-ip>
     ```

#### **Step 5: Use a Systemd Service for Reliability (Optional)**

If you don’t want to use Docker, set up your app as a **systemd service** so it runs indefinitely.

1. Create a systemd service file:

   ```bash
   sudo nano /etc/systemd/system/multichat.service
   ```

2. Add the following configuration:

   ```ini
   [Unit]
   Description=MultiChat SSH Server
   After=network.target

   [Service]
   User=<your-username>
   ExecStart=/path/to/multichat
   Restart=always
   EnvironmentFile=/path/to/.env

   [Install]
   WantedBy=multi-user.target
   ```

3. Start and enable the service:
   ```bash
   sudo systemctl start multichat
   sudo systemctl enable multichat
   ```

#### **6. Security Best Practices**

1. **Use Strong SSH Keys**:

   - Ensure all users generate **strong SSH keys** (e.g., ed25519).

2. **Rate-Limit Connections**:

   - Use tools like `fail2ban` to prevent brute-force attacks:
     ```bash
     sudo apt install fail2ban
     ```

3. **Enable Firewall**:

   - Limit access to port `23234` to trusted IPs (if possible):
     ```bash
     sudo ufw allow from <trusted-ip> to any port 23234
     ```

4. **Run as a Non-Root User**:

   - Avoid running the app as `root` in Docker or directly.

5. **Use HTTPS for Sensitive Data**:
   - If you plan to expand the app to include non-SSH interfaces (like a web frontend), ensure all communication is encrypted.
