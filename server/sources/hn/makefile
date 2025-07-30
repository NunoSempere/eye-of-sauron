MAX_LOG_SIZE=20000

# hackernews
run:
	go run main.go fetch.go types.go filterAndExpandSource.go saveSource.go

listen:
	tail -f v2.log

rotate-data:
	# TODO: rotate postgres stuff
	tail -n $(MAX_LOG_SIZE) sources/hn.log | tee -a sources/hn.log.tmp
	mv sources/hn.log.tmp sources/hn.log


systemd:
	sudo cp hn.service /etc/systemd/system
	sudo systemctl daemon-reload
	sudo systemctl enable potpourri
	sudo systemctl restart potpourri 
