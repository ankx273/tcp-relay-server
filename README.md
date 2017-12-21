# tcp-relay-server
tcp relay server to talk to echo server(s)

How to run this :

1. go build relay-server.go
2. ./relay -port 7000

Now we can start the echo server at the relayed port and get the new public relay address

3. go build echo-server.go
4. ./echo-server -relay localhost:7000

New relay connection estabilished : localhost:5000

5. telnet loclahost 7000 ----> you can start multiple clients at 7000 and you should get your response back
hello,world
>>hello,world

6. You can start another echo server at the same relay address
7./echo-server localhost:7000
New relay connection estabilished : localhost:5001

8. telnet localhost:5001 ----> you can start multiple clients at 7000 and you should get your response back

hello, again
>>hello, again
