MAX_LOG_SIZE=20000

# wikinews
run:
	go run fetchWikinews.go filterAndExpandSource.go main.go

listen:
	tail -f v2.log

rotate: 
	# TODO: rotate postgres stuff
	tail -n $(MAX_LOG_SIZE) v2.log | tee -a v2.log.tmp
	mv v2.log.tmp v2.log

systemd:
	sudo cp wikinews.service /etc/systemd/system
	sudo systemctl daemon-reload
	sudo systemctl enable wikinews 
	sudo systemctl restart wikinews

status:
	systemctl status wikinews --no-pager
