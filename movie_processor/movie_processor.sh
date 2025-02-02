#!/bin/bash

# Reading API key from key.txt file
API_KEY=$(<key.txt)

# Check if the API key is indeed read, otherwise abort
if [ -z "$API_KEY" ]; then
    echo "API key could not be read from key.txt"
    exit 1
fi

# Prompt user for input and output paths
echo "Enter the path to your movie list file (Press Enter to skip):"
read INPUT_PATH

echo "Enter the path for the output CSV file (Press Enter to skip):"
read OUTPUT_PATH

# Building the arguments for python command conditionally
CMD_ARGS="$API_KEY"

if [ -n "$INPUT_PATH" ]; then
    CMD_ARGS+=" --list \"$INPUT_PATH\""
fi

if [ -n "$OUTPUT_PATH" ]; then
    CMD_ARGS+=" --output \"$OUTPUT_PATH\""
fi

# Running the Python script with the provided details
eval "python extend_movie_list.py $CMD_ARGS"

echo "Script execution completed."