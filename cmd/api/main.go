package main

import "log"

func main() {
	cfg := config{
		addr: ":8080",
	}

	app := application{
		config: cfg,
	}

	err := app.serve(app.mountRoutes())
	if err != nil {
		log.Fatal(err)
	}

}
