**некоторые мысли по решению:**
1. По ручке CreateScript() сервис слой создает 
    в бд скрипт с определенной командой, после чего сервис его запускает,
    для того, чтобы вывод команды записался в бд, необходимо запускать скрипт
    в отдельной горутине и в defer обновлять данные в бд

### Links

[How Does Linux deal with shell scripts (StackOverflow)](https://unix.stackexchange.com/questions/121013/how-does-linux-deal-with-shell-scripts)
