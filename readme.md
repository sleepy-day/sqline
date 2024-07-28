# Sqline
A Terminal UI tool for querying and managing SQL databases, [Tcell](https://github.com/gdamore/tcell) is used for creating the Terminal UI. Currently in development.
# Build Instructions
Makefile will be provided for building for other systems or recommended compile flags.
TODO: make the makefile
# Getting started
TODO
# Features
- Text editor (!!)
- list more
# Todo
- Allow the text editor to buffer text and only send it through when necessary
# Feature Explanations
## General
The terminal UI was implemented using the  library.
## Text Editor
It was implemented using a gap buffer. 
- The buffer and cursor position on the screen are tracked independently due to the difference in dimensions (2D cursor position vs 1D position in array)
- Movement up and down will calculate where to move based off previous/next/following line start and/or end positions, it will maintain the column position when moving between rows and do the entire movement in one ```copy()``` to try and speed things up a little.
- The cursor position is tracked by checking the lengths of the lines and will only allow movement within or at the end of text
- The screen will redraw when the text has updated which will also recalculate the line lengths
- It can open text files directly in the case of wanting to load an SQL script to run
