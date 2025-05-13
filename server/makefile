MAX_LOG_SIZE=20000

# galerts
run-galerts:
	 go run sources/galerts/main.go sources/galerts/filterAndExpandSource.go sources/galerts/saveSource.go sources/galerts/fetchGoogleAlerts.go

listen-galerts:
	tail -f sources/galerts/v2.log

rotate-data-galerts: 
	# TODO: rotate postgres stuff
	tail -n $(MAX_LOG_SIZE) sources/galerts/v2.log | tee -a sources/galerts/v2.log.tmp
	mv sources/galerts/v2.log.tmp sources/galerts/v2.log

# gdelt 
run-gdelt:
	 go run sources/gdelt/main.go sources/gdelt/filterAndExpandSource.go sources/gdelt/saveSource.go sources/gdelt/fetchGKG.go

listen-gdelt:
	tail -f sources/gdelt/v2.log

rotate-data-gdelt: 
	# TODO: rotate postgres stuff
	tail -n $(MAX_LOG_SIZE) sources/gdelt/v2.log | tee -a sources/gdelt/v2.log.tmp
	mv sources/gdelt/v2.log.tmp sources/gdelt/v2.log

# wikinews
run-wikinews:
	go run sources/wikinews/fetchWikinews.go sources/wikinews/filterAndExpandSource.go sources/wikinews/saveSource.go sources/wikinews/main.go

listen-wikinews:
	tail -f sources/wikinews/v2.log

rotate-data-wikinews: 
	# TODO: rotate postgres stuff
	tail -n $(MAX_LOG_SIZE) sources/wikinews/v2.log | tee -a sources/wikinews/v2.log.tmp
	mv sources/wikinews/v2.log.tmp sources/wikinews/v2.log

# gmw
run-gmw:
	go run sources/gmw/mil/fetchGmwMil.go sources/gmw/mil/filterAndExpandSource.go sources/gmw/mil/main.go sources/gmw/mil/types.go sources/gmw/mil/saveSource.go

listen-gmw:
	tail -f sources/gmw/mil/v2.log

rotate-data-gmw: 
	# TODO: rotate postgres stuff
	tail -n $(MAX_LOG_SIZE) sources/gmw/mil/v2.log | tee -a sources/gmw/mil/v2.log.tmp
	mv sources/gmw/mil/v2.log.tmp sources/gmw/mil/v2.log

# potpourri
run-potpourri:
	go run sources/potpourri/main.go sources/potpourri/filterAndExpandSource.go sources/potpourri/saveSource.go 

listen-potpourri:
	tail -f sources/potpourri/v2.log

rotate-data-potpourri:
	# TODO: rotate postgres stuff
	tail -n $(MAX_LOG_SIZE) sources/potpourri/v2.log | tee -a sources/potpourri/v2.log.tmp
	mv sources/potpourri/v2.log.tmp sources/potpourri/v2.log

# hackernews
run-hn:
	go run sources/potpourri/hn/main.go sources/potpourri/hn/fetch.go sources/potpourri/hn/types.go sources/potpourri/hn/filterAndExpandSource.go sources/potpourri/hn/saveSource.go

listen-hn:
	tail -f sources/potpourri/hn.log

rotate-data-hn:
	# TODO: rotate postgres stuff
	tail -n $(MAX_LOG_SIZE) sources/potpourri/hn.log | tee -a sources/potpourri/hn.log.tmp
	mv sources/potpourri/hn.log.tmp sources/potpourri/hn.log

# Others
deps:
	go mod tidy
	go mod download
	go mod vendor

systemd: systemd/*
	sudo cp systemd/galerts.service /etc/systemd/system
	sudo cp systemd/gdelt.service /etc/systemd/system
	sudo cp systemd/wikinews.service /etc/systemd/system
	sudo cp systemd/gmw.service /etc/systemd/system
	sudo cp systemd/potpourri.service /etc/systemd/system
	sudo cp systemd/hn.service /etc/systemd/system
	sudo systemctl daemon-reload
	sudo systemctl enable galerts
	sudo systemctl restart galerts
	sudo systemctl restart gdelt
	sudo systemctl restart wikinews
	sudo systemctl restart gmw
	sudo systemctl restart potpourri
	sudo systemctl restart hn
