A very simple logic gate simulator.

# Setup
Enter the [Nix](https://nixos.org/) shell
```
nix develop
```
Run the program
```
go run src/main.go
```

Note: all the code is located in a single file that isn't organized very well.

# Usage
Build a logic gate network by dragging gates onto the canvas. Delete they using right-click. Connect gates by click-and-dragging from their sockets. Go to the simulation mode by pressing `enter`. Step one tick forward using the `step` button. From this screen switches can by toggled by clicking on them.

Note: I couldn't be bothered to properly implement the deletion of gates so crashes are very likely.
