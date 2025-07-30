MAX_LOG_SIZE=20000

run:
	 go run main.go filterAndExpandSource.go saveSource.go fetchGoogleAlerts.go
	
listen:
	tail -f v2.log

rotate:
	tail -n $(MAX_LOG_SIZE) v2.log | tee -a v2.log.tmp
	mv v2.log.tmp v2.log

systemd:
	sudo cp galerts.service /etc/systemd/system
	sudo systemctl daemon-reload
	sudo systemctl enable galerts
	sudo systemctl restart galerts
