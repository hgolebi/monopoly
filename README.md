## 🚀 How to Run the Game

Follow these steps to set up and run the game server and clients.

---

### I. Prerequisites: Install and Set Up Go

1.  **Install GO Language:**
    * Download and install Go version 1.23 or newer from the official Go website: [https://go.dev/doc/install](https://go.dev/doc/install)
2.  **Add Go to Your PATH:**
    * Ensure the Go binary directory is added to your system's **PATH** environment variable if the installer hasn't done it automatically.

---

### II. Set Up Project Dependencies

3.  **Download Dependencies:**
    * Navigate to the main project files folder in your terminal.
    * Run the following command to download all necessary project dependencies:
        ```bash
        go mod tidy
        ```

---

### III. Start the Game

4.  **Start the Game Server:**
    * From the main project files folder, run the server process:
        ```bash
        go run main.go
        ```
5.  **Input Player Count:**
    * The server console will prompt you to **input the number of human players**. Enter the required number and press Enter.

6.  **Start Console Clients:**
    * For **every human player**, open a **new console/terminal window** and run the client command:
        ```bash
        go run main.go --cli
        ```

7.  **Play the Game:**
    * Follow the **instructions** displayed in each console client window to play the game.