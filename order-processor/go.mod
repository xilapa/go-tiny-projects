module github.com/xilapa/go-tiny-projects/order-processor

go 1.19

require (
	github.com/google/uuid v1.3.0
	github.com/mattn/go-sqlite3 v1.14.16
	github.com/rabbitmq/amqp091-go v1.5.0
	github.com/xilapa/go-tiny-projects/strong-rabbit v0.0.0-00010101000000-000000000000
	github.com/xilapa/go-tiny-projects/test-assertions v0.0.0-00010101000000-000000000000
)

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/ilyakaznacheev/cleanenv v1.4.2 // indirect
	github.com/joho/godotenv v1.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	olympos.io/encoding/edn v0.0.0-20201019073823-d3554ca0b0a3 // indirect
)

replace github.com/xilapa/go-tiny-projects/test-assertions => ../test-assertions

replace github.com/xilapa/go-tiny-projects/strong-rabbit => ../strong-rabbit
