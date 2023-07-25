all:
	go build -o npuzzle

clean:
	rm -f npuzzle

re: clean all
