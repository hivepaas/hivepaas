package kopia

// AllCompressionAlgorithms are the compression algorithms a repository accepts. Anything else is
// rejected by the engine, so it is worth catching before the request reaches it.
var AllCompressionAlgorithms = []string{
	"none",
	"deflate-best-compression", "deflate-best-speed", "deflate-default",
	"gzip", "gzip-best-compression", "gzip-best-speed",
	"lz4",
	"pgzip", "pgzip-best-compression", "pgzip-best-speed",
	"s2-better", "s2-default", "s2-parallel-4", "s2-parallel-8",
	"zstd", "zstd-best-compression", "zstd-better-compression", "zstd-fastest",
}
