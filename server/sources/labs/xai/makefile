build:
	go build -o xai

run: build
	./xai

install: build
	sudo cp xai /usr/local/bin/
	sudo cp xai.service /etc/systemd/system/
	sudo systemctl daemon-reload
	sudo systemctl enable xai.service

start:
	sudo systemctl start xai.service

stop:
	sudo systemctl stop xai.service

status:
	sudo systemctl status xai.service

logs:
	sudo journalctl -u xai.service -f

clean:
	rm -f xai

.PHONY: build run install start stop status logs clean