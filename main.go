package main

func main() {
	blockChain := NewBlockChain()
	cli := CLI{blockChain}
	cli.Run()
}
