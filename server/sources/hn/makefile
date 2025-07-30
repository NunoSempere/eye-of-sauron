MAX_LOG_SIZE=20000

# hackernews
run:
	go run main.go fetch.go types.go filterAndExpandSource.go saveSource.go

listen:
	tail -f v2.log

rotate:
	# TODO: rotate postgres stuff
	tail -n $(MAX_LOG_SIZE) v2.log | tee -a v2.log.tmp
	mv v2.log.tmp v2.log


systemd:
	sudo cp hn.service /etc/systemd/system
	sudo systemctl daemon-reload
	sudo systemctl enable hn
	sudo systemctl restart hn 
