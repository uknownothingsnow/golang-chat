Try to write a chat server use Golang.

go run chat.go  will start chat server on port 6000

use telnet localhost 6000 to start a client, after client connected successfully, server will send auth to client, and wait for client to input a nickname, this nickname is then used as communication id.

sent text with format nickname:message.

for example a:hello will send hello to user a.

I am working to make this better, enjoy it.