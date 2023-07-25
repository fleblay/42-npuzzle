all:
	go build .

clean:
	rm -f npuzzle

re: clean all
