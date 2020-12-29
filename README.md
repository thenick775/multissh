# multissh
A simple TUI utility to manage multiple terminals at once with synchronization
<div align="center">  
 
<img src="https://github.com/thenick775/multissh/blob/main/graphics/demo.gif" width="70%" >

</div>

### Run Instructions
Put your pem file locations, server addresses, and usernames in the `serverloc.txt` file (or any other file) and then run
```
go run main.go serverloc.txt
```

### Usage
 - Use F1 to toggle synchronization of the multiple connections, meaning whether commands execute on all connectiosn or only the one currently selected
 - Use the 'loadCommand(<filename>)' trigger to load a command from a file (done as copy/paste is non functional in the go-tui input). Absolute or local path, no quotes
 - Use Esc to quit and disconnect
