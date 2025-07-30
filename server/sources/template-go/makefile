MAX_LOG_SIZE=20000

# {{SOURCE_NAME}}
run:
	go run fetch.go filterAndExpandSource.go saveSource.go main.go

listen:
	tail -f v2.log

rotate: 
	# TODO: rotate postgres stuff
	tail -n $(MAX_LOG_SIZE) v2.log | tee -a v2.log.tmp
	mv v2.log.tmp v2.log

systemd:
	sudo cp {{SOURCE_NAME}}.service /etc/systemd/system
	sudo systemctl daemon-reload
	sudo systemctl enable {{SOURCE_NAME}}
	sudo systemctl restart {{SOURCE_NAME}}