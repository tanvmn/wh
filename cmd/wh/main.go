package main

type config struct {
	port string
	env  string
	dsn  string
}

type application struct {
	cfg config
}

func main() {
	println("hello from Go")
}
