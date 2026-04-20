go build -tags testbuild -o ti . && go test ./test/... -count=1 -parallel=4 && rm ti
