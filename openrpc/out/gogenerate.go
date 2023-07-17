package out

//go:generate go run ./../cmd compile -s api-spec/schemas -m api-spec/methods -o spec.json
//go:generate go run  ./../cmd generate -p out -s spec.json -o generated_api.go -t ../templates/types.gotmpl
