# Swagen

**Features**

1. Merge swagger files into one swagger file
2. Generate React-Redux-TS client code


merge
```
go-swagen merge -c 1 -p \
  -i account@./build/inputs/account.swagger.json \
  -i file@./build/inputs/file.swagger.json \
  -i finance@./build/inputs/finance.swagger.json \
  -i retail@./build/inputs/retail.swagger.json \
  -i stock@./build/inputs/stock.swagger.json \
  -o ./build/gen/swagger.json
```

generate
```
go-swagen generate ./build/gen/swagger.json -o ./build/gen
```


# get go releaser binary
```
curl -sL https://git.io/goreleaser | bash
```