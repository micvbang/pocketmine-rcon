# Introduction
This is a very simple, synchronous Minecraft and Minecraft Pocket Edition (Minecraft PE, Pocketmine) RCON client, implemented to [these specs](http://wiki.vg/Rcon).


[cli/main.go](cli/main.go) implements a REPL CLI as an example of how to use the client.

# Shortcomings
- Synchronous.
- Long (split) responses from server aren't handled correctly.
- Probably a lot more.


# License
See the [LICENSE](LICENSE) file.

