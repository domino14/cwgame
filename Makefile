pb:
	mkdir -p gen;
	protoc --go_out=gen --go_opt=paths=source_relative ./proto/cwgame/cwgame.proto

clean:
	rm -f bin/*
