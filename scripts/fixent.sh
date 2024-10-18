cd .. 
go run -mod=mod entgo.io/ent/cmd/ent generate ./ent/schema
go clean -modcache
go mod tidy
# go get -u entgo.io/ent