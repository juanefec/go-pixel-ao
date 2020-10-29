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

1. ``cd go-pixel-ao/client``
2. ``go run .``


## Client Logic

- Load images
- instanciate spells player & connect 2 sv
- start window
- start listening input
- start game update loop
 - game update is always sending an update/event/etc message and reciving new info of players each time it sends 
- start draw update loop
 - draw things in the less annoying way to see it
 - possible solve to draw correctly, divide batch by position (downside: always re-arranging it when we move)


## Wanted features xd:
- dmg info when reciving (example: -110(in red for dmg)), also animate it in a way that when it appears its big and it rises and turns smaller :/

- make a dummie to hit

- god panel :B

- obiously better collisions and an order to draw things correctly, but those hard :S