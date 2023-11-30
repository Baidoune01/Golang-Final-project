# Key-Value Store System

## Overview
This Key-Value Store System is designed as an API server in Go, implementing LSM-tree (Log-Structured Merge-tree) concepts for efficient write and read operations. It consists of several components, including a MemTable, Write-Ahead Log (WAL), SSTable (Sorted String Table) Manager, and a Compaction Manager. The system provides `set`, `get`, and `delete` operations for key-value pairs and includes a script for easier interaction with the server.

## Components
- **MemTable**: An in-memory data structure to store key-value pairs temporarily before they are flushed to disk.
- **WAL (Write-Ahead Log)**: Records all write operations to ensure data persistence and recovery in case of system failure.
- **SSTable Manager**: Manages the SSTables where the key-value pairs are stored on the disk.
- **Compaction Manager**: Periodically compacts SSTables to reduce data redundancy and improve read efficiency.

## Installation
 1. Navigate to the project directory and run the server:
   ```
   go run .
   ```

## Usage
The server listens on `localhost:8080` and supports the following endpoints:

- **Set a key-value pair**:  
  `POST /set` with a JSON body `{"key": "value"}`

- **Get a value by key**:  
  `GET /get?key=<key>`

- **Delete a key-value pair**:  
  `DELETE /del?key=<key>`

## Command-Line Interface (CLI)
For easier interaction with the Key-Value Store System, a Bash script `KvStoreCLI.sh` is provided. This script offers a simple command-line interface to perform `set`, `get`, and `delete` operations.

### Using KvStoreCLI.sh
1. Ensure you have `curl` installed on your system.
2. Run the script:
   ```
   bash KvStoreCLI.sh
   ```
3. You will see a prompt `> `. Here, you can enter commands in the format:
   - To set a key-value pair: `set <key> <value>`
   - To get a value by key: `get <key>`
   - To delete a key-value pair: `del <key>`
   - To exit the script: `exit`

## Examples
1. **Setting a key-value pair**: 
   ```
   > set foo bar
   ```

2. **Getting a value by key**:
   ```
   > get foo
   bar
   ```

3. **Deleting a key-value pair**:
   ```
   > del foo
   Key deleted successfully
   ```

4. **Exiting the CLI**:
   ```
   > exit
   ```

## Compaction
The system automatically triggers compaction after every four SSTables are created. Compaction merges these SSTables, removes redundant data, and keeps only the latest version of each key.
