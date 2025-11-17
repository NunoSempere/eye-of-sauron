MAX_LOG_SIZE=20000

# gmw
# potpourri
run:
	go run main.go filterAndExpandSource.go 
listen:
	tail -f v2.log

rotate:
	# TODO: rotate postgres stuff
	tail -n $(MAX_LOG_SIZE) v2.log | tee -a v2.log.tmp
	mv v2.log.tmp v2.log

systemd:
	sudo cp potpourri.service /etc/systemd/system
	sudo systemctl daemon-reload
	sudo systemctl enable potpourri
	sudo systemctl restart potpourri 

status:
	systemctl status potpourri --no-pager
