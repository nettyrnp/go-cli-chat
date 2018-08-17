# Application for chatting via command line interface

Features:

* High throughput
* Timeout on user idleness
* Server notifies all chat members when some user comes online or goes offline
* Each message is broadcasted to all chat members

## How to launch the application

Install the application from the Terminal:
```shell
go get github.com/nettyrnp/go-cli-chat
```

Build the application from the Terminal or from under IntelliJ IDEA etc.:
```shell
go build main.go
```

Run the application from the Terminal (or in Windows Explorer double click on file 'main.exe'):
```shell
./main.exe
```

Run another Terminal, then run telnet in it:
```shell
# Terminal:
Добро пожаловать в программу-клиент Microsoft Telnet
Символ переключения режима: 'CTRL+]'
Microsoft Telnet> o localhost 50505
Подключение к localhost...
Microsoft Telnet>
```

Type the following command:
```shell
# Microsoft Telnet>:
o localhost 50505
```

Type the following words and see them broadcasted to you back:
```shell
# Microsoft Telnet>:
>> Please type in your name
                           Alice
*** Alice is online
Hi Bob
Alice:  Hi Bob
Hi again
Alice:  Hi again
And again
Alice:  And again
Hello everybody
```

Run one or more other Terminals and join the chat under any name:
```shell
# Microsoft Telnet>:
>> Please type in your name
                           Bob
*** Bob is online
*** Alice is online
Alice:      Hi
Alice:  Hi hi
Alice:  Hello hello
*** Alice is offline
*** Alice is online
```

