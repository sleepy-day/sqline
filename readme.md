# Sqline
A Terminal UI tool for querying and managing SQL databases, [Tcell](https://github.com/gdamore/tcell) is used for creating the terminal UI. 
# Build Instructions
A makefile has been provided for building the executable for Linux/Windows though I haven't tested the command on Windows to see if it will work, released built this way will be put in the folders ```release/linux``` and ```release/windows```. It can also be compiled by running ```go build``` though CGO is required so a C compiler will need to be installed.
# Project Structure
- app
  - Set up for the main program itself and where everything is called from
- components
  - Contains the UI elements, these are reused throughout the project
- util
  - Contains the code for saving/loading the config and the gap buffer code
- db
  - Contains implementations for database interfaces (Currently has a minimal Sqlite and very partial Postgres implementation)
- views
  - Contains the different views which use components to make up different screens/menus
# Features
- Works with Sqlite and can be expanded to others through the use of an interface
- Custom terminal UI components such as:
  - Text Editor
  - Lists
  - Tree Lists
  - Tables
  - Forms
  - Textbox
  - Radio Selectors
  - Buttons
  - Status Bar
- Gap buffer implementation for the text editor
- Saving and loading connections to and from a config file
  - Will save any connections saved within the program to the config dir based on your OS from the ```os.UserConfigDir``` function, keep this in mind if running the program in case you don't want it saved locally
- Displays Tables and their columns, data from queries, results from updates/inserts and indexes and their attributes
# Showcase
![Insert](https://github.com/user-attachments/assets/ee144dfc-6480-470a-9250-cc4cf81bc6a0)
![Index](https://github.com/user-attachments/assets/ce5cd01c-d7fd-41f9-8192-00ae4821ebd7)
![CreateTableAndOpenList](https://github.com/user-attachments/assets/90e9bbe6-1e9a-4810-970f-3fb65449e43f)
![CreateMultipleTables](https://github.com/user-attachments/assets/6b3620e5-5ce4-4d44-8cb1-60eddf840cca)
![CreateAndConnect](https://github.com/user-attachments/assets/91157e39-7005-428e-8585-bf48d213d902)
