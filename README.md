# Swagen

**Features**

1. Merge swagger files into one swagger file
2. Generate React-Redux-TS client code


merge
```
go run cmd/swagen.go merge -c 1 -p \
  -i account@./build/account.swagger.json \
  -i catalog@./build/catalog.swagger.json \
  -i file@./build/file.swagger.json \
  -i finance@./build/finance.swagger.json \
  -i merchant@./build/merchant.swagger.json \
  -i pingpp@./build/pingpp.swagger.json \
  -i retail@./build/retail.swagger.json \
  -i stock@./build/stock.swagger.json \
  -o ./build/gen/swagger.json
```

generate
```
go run cmd/swagen.go generate ./build/gen/swagger.json -o ./build/gen
```

filter
```
// -i input
// -t tags
// -p pretty
// -o output
go run cmd/swagen.go filter \
  -i ./build/swagger-input.json \
  -o ./build/swagger.json \
  -t "AccountNSService" \
  -t "CatalogSystemService" \
  -p
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

# goreleaser
```
VERSION=0.6.0 && git tag -a v${VERSION} -m "release v${VERSION}" && git push origin --follow-tags
goreleaser --rm-dist
```