#!/bin/bash

# Function to send a SET request to the server
set_key() {
    response=$(curl -s -X POST -H "Content-Type: application/json" -d "{\"$1\":\"$2\"}" http://localhost:8080/set)
    echo "$response"
}

# Function to send a GET request to the server
get_key() {
    response=$(curl -s -X GET http://localhost:8080/get?key=$1)
    echo "$response"
}

# Function to send a DELETE request to the server
delete_key() {
    response=$(curl -s -X DELETE http://localhost:8080/del?key=$1)
    echo "$response"
}

# Main loop for the command-line interface
while true; do
    echo -n "> "
    read -r cmd key value

    case $cmd in
        set)
            set_key "$key" "$value"
            ;;
        get)
            get_key "$key"
            ;;
        del)
            delete_key "$key"
            ;;
        exit)
            break
            ;;
        *)
            echo "Unknown command. Available commands: set, get, del, exit"
            ;;
    esac
done
