package meta

type Config struct {
	Addr      string
	Password  string
	IORetries int
}

type Format struct {
	Name        string
	UUID        string
	Storage     string
	Bucket      string
	AccessKey   string
	SecretKey   string
	BlockSize   int
	Compression string
	Partitions  int
}
