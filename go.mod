module github.com/wispberry-technologies/wispy-core

go 1.23.7

replace github.com/wispberry-technologies/wispy-core/common => ./common

replace github.com/wispberry-technologies/wispy-core/auth => ./auth

replace github.com/wispberry-technologies/wispy-core/cache => ./cache

replace github.com/wispberry-technologies/wispy-core/api => ./api

require (
	github.com/glebarez/go-sqlite v1.22.0
	github.com/go-chi/chi/v5 v5.2.1
	github.com/go-chi/httprate v0.15.0
	github.com/joho/godotenv v1.5.1
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/uuid v1.5.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	golang.org/x/sys v0.33.0 // indirect
	modernc.org/libc v1.37.6 // indirect
	modernc.org/mathutil v1.6.0 // indirect
	modernc.org/memory v1.7.2 // indirect
	modernc.org/sqlite v1.28.0 // indirect
)
