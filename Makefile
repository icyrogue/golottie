coverage:
	go test -cover -coverprofile coverage.out

covsh: coverage
	GOCOVSH_THEME=macchiato gocovsh 

benchmark:	
	go test -bench=. -run=^# ./...