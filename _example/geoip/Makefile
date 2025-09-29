SHELL=/bin/bash

run:
	go run main.go --config=config.yaml

check:
	curl  -H "X-Real-IP: 1.1.1.1" http://127.0.0.1:10000/ && echo "" && echo "" && echo ""
	curl  -H "CF-IPCountry: UNK" http://127.0.0.1:10000/ && echo "" && echo "" && echo ""