# multissh
A simple TUI utility to manage multiple terminals at once with synchronization.

I end up doing a lot of tasks spread upon multiple servers using scp/ssh, this utility was designed to make the ssh portion of my bulk actions faster with less clutter.
<div align="center">  
 
<img src="https://github.com/thenick775/multissh/blob/main/graphics/demo.gif" width="70%" >

</div>

### Run Instructions
Put your pem file locations, server addresses, and usernames in the `serverloc.txt` file (or any other file) and then run
```
go run main.go serverloc.txt
```

### Usage
 - Use Ctrl+s to toggle synchronization of the multiple connections, meaning whether commands execute on all connectiosn or only the one currently selected
 - Use the 'loadCommand(<filename>)' trigger to load a command from a file (done as copy/paste is non functional in the go-tui input). Absolute or local path, no quotes
 - Use Ctrl+t to quick scroll to the top
 - Use Ctrl+b to quick scroll to the bottom
 - Use up and down arrow keys to scroll
 - Use Esc to quit and disconnect
