@echo off
setlocal enabledelayedexpansion

REM Reading API Key from key.txt
set /p API_KEY=<key.txt

REM Prompting user for input file path and output file path
set /p INPUT_PATH=Enter the path of the input movie list file (press Enter to use default):
if not "!INPUT_PATH!"=="" set INPUT_ARG=--list "!INPUT_PATH!"
else set INPUT_ARG=

REM Prompting user for output file path
set /p OUTPUT_PATH=Enter the path of the output CSV file (press Enter to use default):
if not "!OUTPUT_PATH!"=="" set OUTPUT_ARG=--output "!OUTPUT_PATH!"
else set OUTPUT_ARG=

REM Running the Python script with provided inputs
python extend_movie_list.py !API_KEY! !INPUT_ARG! !OUTPUT_ARG!

echo Script execution completed
pause