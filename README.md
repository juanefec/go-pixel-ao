# Go AO

Trying to develop some sort of game using only go and a basic tcp4 socket connection

I'm using [pixel](https://github.com/faiface/pixel "Pixel Github") for drawing and interacting with the user system

![](demo.gif)

## How to run
### Server
1. ``git clone https://github.com/juanefec/go-pixel-ao``
2. ``cd go-pixel-ao``
3. ``go run ./server/main.go``
### Client

1. ``go run ./client/main.go ./client/utils.go ./client/keys.go``