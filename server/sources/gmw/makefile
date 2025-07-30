MAX_LOG_SIZE=20000

run:
	go run mil/fetchGmwMil.go mil/filterAndExpandSource.go mil/main.go mil/types.go mil/saveSource.go

listen:
	tail -f mil/v2.log

rotate: 
	# TODO: rotate postgres stuff
	tail -n $(MAX_LOG_SIZE) mil/v2.log | tee -a mil/v2.log.tmp
	mv mil/v2.log.tmp mil/v2.log

systemd:
	sudo cp gdelt.service /etc/systemd/system
	sudo systemctl daemon-reload
	sudo systemctl enable gmw
	sudo systemctl restart gmw 
