# Gator CLI - RSS Feed Aggregator

Gator CLI is a command-line application written in Go for managing and aggregating RSS feeds. It allows users to register, manage multiple feeds, follow/unfollow specific feeds stored in a PostgreSQL database, and browse aggregated posts fetched from those feeds.

## Features

* **User Management:**
    * Register new users.
    * Log in as a specific user (sets the active user for subsequent commands).
    * List all registered users.
    * Reset the database (removes all users, feeds, and posts - **Use with caution!**).
* **Feed Management:**
    * Add new RSS feeds with a name and URL.
    * List all feeds stored in the database.
    * Follow existing feeds.
    * List feeds followed by the current user.
    * Unfollow feeds.
* **Aggregation:**
    * Periodically fetch new posts from registered feeds.
* **Browse:**
    * View posts fetched from followed feeds.

## Prerequisites

* **Go:** Ensure you have Go installed (version 1.18 or later recommended).
* **PostgreSQL:** A running PostgreSQL database instance.
* **Database Schema:** The necessary database tables must be created. The schema definitions are expected to be compatible with the queries defined in the `internal/database` package (likely generated using a tool like `sqlc` from SQL files located elsewhere in the project).

## Installation & Setup

1.  **Clone the Repository:**
    ```bash
    git clone <your-repository-url> # Replace with the actual URL
    cd gator # Or your repository's directory name
    ```

2.  **Database Setup:**
    * Connect to your PostgreSQL instance.
    * Create a database for the application (e.g., `gator_db`).
    * Create a user and grant privileges on the database (e.g., `gator_user`).
    * Apply the required database schema. You'll need the SQL files that define the tables (`users`, `feeds`, `feed_follows`, `posts`) corresponding to the `internal/database` package.

3.  **Configuration:**
    * The application uses the `internal/config` package to manage configuration. This typically involves setting environment variables or creating a configuration file.
    * The most crucial configuration is the **Database Connection URL (`DBURL`)**. It should be in a format like:
        ```
        postgres://<user>:<password>@<host>:<port>/<dbname>?sslmode=disable
        ```
        Replace the placeholders with your actual database credentials.
    * Refer to the implementation of the `internal/config` package for specific details on how it loads configuration (e.g., expected file names, environment variable names).

4.  **Build the Application:**
    ```bash
    go build -o gator .
    ```
    This will create an executable file named `gator` (or `gator.exe` on Windows) in the current directory.

## Usage

Run the application from your terminal using the built executable:

```bash
./gator <command> [arguments...]