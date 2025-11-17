MAX_LOG_SIZE=20000

# Anthropic News
run:
	go run fetch.go filterAndExpandSource.go  main.go

listen:
	tail -f v2.log

rotate: 
	# TODO: rotate postgres stuff
	tail -n $(MAX_LOG_SIZE) v2.log | tee -a v2.log.tmp
	mv v2.log.tmp v2.log

systemd:
	sudo cp anthropic.service /etc/systemd/system
	sudo systemctl daemon-reload
	sudo systemctl enable anthropic
	sudo systemctl restart anthropic 

status:
	systemctl status anthropic --no-pager
