go run ./cmd/swagen.go merge -c 1 -p \
  -i account@./build/inputs/account.swagger.json \
  -i file@./build/inputs/file.swagger.json \
  -i finance@./build/inputs/finance.swagger.json \
  -i retail@./build/inputs/retail.swagger.json \
  -i stock@./build/inputs/stock.swagger.json \
  -o ./build/gen/swagger.json