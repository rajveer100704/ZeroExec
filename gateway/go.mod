module github.com/rajveer100704/ZeroExec/gateway

go 1.26.1

require (
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/gorilla/websocket v1.5.3
	github.com/joho/godotenv v1.5.1
	github.com/rajveer100704/ZeroExec/agent v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v3 v3.0.1
)

require golang.org/x/sys v0.42.0 // indirect

replace github.com/rajveer100704/ZeroExec/agent => ../agent
