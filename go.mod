module github.com/wispberry-technologies/wispy-core

go 1.23.7

replace github.com/wispberry-technologies/wispy-core/common => ./common
replace github.com/wispberry-technologies/wispy-core/auth => ./auth
replace github.com/wispberry-technologies/wispy-core/cache => ./cache
replace github.com/wispberry-technologies/wispy-core/api => ./api

require (
	github.com/go-chi/chi/v5 v5.2.1
	github.com/go-chi/httprate v0.15.0
	github.com/joho/godotenv v1.5.1
)

require (
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	golang.org/x/sys v0.33.0 // indirect
)
