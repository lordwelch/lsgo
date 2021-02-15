module git.narnian.us/lordwelch/lsgo

go 1.15

replace github.com/pierrec/lz4/v4 v4.1.3 => ./third_party/lz4

require (
	github.com/go-git/go-git/v5 v5.2.0
	github.com/go-kit/kit v0.10.0
	github.com/google/uuid v1.1.4
	github.com/kr/pretty v0.2.1
	github.com/pierrec/lz4/v4 v4.1.3
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a
	golang.org/x/sys v0.0.0-20210108172913-0df2131ae363
	gonum.org/v1/gonum v0.8.2
)
