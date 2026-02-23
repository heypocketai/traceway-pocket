package main

import tracewaybackend "github.com/tracewayapp/traceway/backend"

func main() {
	tracewaybackend.Run(
		tracewaybackend.WithPort(8082),
		tracewaybackend.WithDefaultUser("admin@localhost.com", "admin"),
		tracewaybackend.WithDefaultProject("Backend", "go", "backend-dev-token"),
		tracewaybackend.WithDefaultProject("Frontend", "sveltekit", "frontend-dev-token"),
	)
}
