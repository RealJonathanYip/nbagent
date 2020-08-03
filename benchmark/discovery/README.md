## Windows
### register
discovery.exe --agents=127.0.0.1:8800 --call-timeout=10 --clients=1 --only-agent=true register --secret-key=node --total-register=1 --uri-size=256

### read
discovery.exe --agents=127.0.0.1:8800 --call-timeout=10 --clients=1 --only-agent=true read --secret-key=node --total-read=1 --uri-size=256



## Linux
### register
./discovery --agents=127.0.0.1:8800 --call-timeout=10 --clients=1 --only-agent=true register --secret-key=node --total-register=10000 --uri-size=256

### read
./discovery --agents=127.0.0.1:8800 --call-timeout=10 --clients=1 --only-agent=true read --secret-key=node --total-read=10000 --uri-size=256