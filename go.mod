module github.com/tracklines/automation

go 1.26.1

// ponytail: until the backend's retention package is pushed, build against the local checkout; then
// `go get github.com/tracklines/backend@main` and delete this line.
replace github.com/tracklines/backend => ../backend

require (
	github.com/bugfixes/go-bugfixes v0.19.5
	github.com/jackc/pgx/v5 v5.11.0
	github.com/tracklines/backend v0.0.0
)

require (
	github.com/go-logfmt/logfmt v0.6.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.3 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/text v0.43.0 // indirect
)
