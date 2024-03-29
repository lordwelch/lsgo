module git.narnian.us/lordwelch/lsgo

go 1.15

replace github.com/pierrec/lz4/v4 v4.1.3 => ./third_party/lz4

require (
	github.com/go-kit/kit v0.10.0
	github.com/google/uuid v1.1.4
	github.com/kr/pretty v0.2.1
	github.com/pierrec/lz4/v4 v4.1.3
	gonum.org/v1/gonum v0.8.2
)
