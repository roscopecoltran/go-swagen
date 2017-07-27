# Swagen

**Features**

1. Merge swagger files into one swagger file
2. Generate React-Redux-TS client code


merge
```
swagen merge -c 1 -p \
  -i account@./build/account.swagger.json \
  -i file@./build/file.swagger.json \
  -i finance@./build/finance.swagger.json \
  -i retail@./build/retail.swagger.json \
  -i stock@./build/stock.swagger.json \
  -o ./build/gen/swagger.json
```

generate
```
swagen generate ./build/gen/swagger.json -o ./build/gen
```


# get go releaser binary
```
curl -sL https://git.io/goreleaser | bash
```

# go-bindata
```
# cd react_redux_typescript
go-bindata templates/
```