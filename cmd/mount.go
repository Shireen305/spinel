package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/VividCortex/godaemon"
	"github.com/google/gops/agent"
	"github.com/sirupsen/logrus"
	"github.com/souvikdeyrit/spinel/pkg/chunk"
	"github.com/souvikdeyrit/spinel/pkg/fuse"
	"github.com/souvikdeyrit/spinel/pkg/meta"
	"github.com/souvikdeyrit/spinel/pkg/redis"
	"github.com/souvikdeyrit/spinel/pkg/utils"
	"github.com/souvikdeyrit/spinel/pkg/vfs"
	"github.com/urfave/cli/v2"
)

func MakeDaemon() {
	godaemon.MakeDaemon(&godaemon.DaemonAttr{})
}

func installHandler(mp string) {
	// Catch all syscalls from OS
	signal.Ignore(syscall.SIGPIPE)
	signalChan := make(chan os.Signal, 10)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	go func() {
		for {
			<-signalChan
			go func() {
				if runtime.GOOS == "linux" {
					exec.Command("umount", mp, "-l").Run()
				} else if runtime.GOOS == "darwin" {
					exec.Command("diskutil", "umount", "force", mp).Run()
				}
			}()
			go func() {
				time.Sleep(time.Second * 3)
				os.Exit(1)
			}()
		}
	}()
}

func mount(c *cli.Context) error {
	/*
		Setup mountpoint
	*/
	go func() {
		for port := 6060; port < 6100; port++ {
			http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil)
		}
	}()
	go func() {
		for port := 6070; port < 6100; port++ {
			agent.Listen(agent.Options{Addr: fmt.Sprintf("127.0.0.1:%d", port)})
		}
	}()
	if c.Bool("trace") {
		utils.SetLogLevel(logrus.TraceLevel)
	} else if c.Bool("debug") {
		utils.SetLogLevel(logrus.DebugLevel)
	} else if c.Bool("quiet") {
		utils.SetLogLevel(logrus.ErrorLevel)
		utils.InitLoggers(!c.Bool("nosyslog"))
	}
	if c.Args().Len() < 1 {
		logger.Fatalf("Redis URL and mountpoint are required")
	}
	addr := c.Args().Get(0)
	if !strings.Contains(addr, "://") {
		addr = "redis://" + addr
	}
	if c.Args().Len() < 2 {
		logger.Fatalf("MOUNTPOINT is required")
	}
	mp := c.Args().Get(1)
	if !utils.Exists(mp) {
		if err := os.MkdirAll(mp, 0777); err != nil {
			logger.Fatalf("create %s: %s", mp, err)
		}
	}

	logger.Infof("Meta address: %s", addr)
	var rc = redis.RedisConfig{Retries: 10, Strict: true}
	m, err := redis.NewRedisMeta(addr, &rc)
	if err != nil {
		logger.Fatalf("Meta: %s", err)
	}
	format, err := m.Load()
	if err != nil {
		logger.Fatalf("load setting: %s", err)
	}

	chunkConf := chunk.Config{
		BlockSize: format.BlockSize * 1024,
		Compress:  format.Compression,

		GetTimeout:  time.Second * time.Duration(c.Int("getTimeout")),
		PutTimeout:  time.Second * time.Duration(c.Int("putTimeout")),
		MaxUpload:   c.Int("maxUpload"),
		AsyncUpload: c.Bool("writeback"),
		Prefetch:    c.Int("prefetch"),
		BufferSize:  c.Int("bufferSize") << 20,

		CacheDir:       c.String("cacheDir"),
		CacheSize:      int64(c.Int("cacheSize")),
		FreeSpace:      float32(c.Float64("freeRatio")),
		CacheMode:      os.FileMode(0600),
		CacheFullBlock: !c.Bool("partialOnly"),
		AutoCreate:     true,
	}

	// Create Redis storage
	blob, err := createStorage(format)
	if err != nil {
		logger.Fatalf("object storage: %s", err)
	}
	logger.Infof("Data use %s", blob)
	logger.Infof("mount volume %s at %s", format.Name, mp)

	if c.Bool("d") {
		MakeDaemon()
	}

	store := chunk.NewCachedStore(blob, chunkConf)
	m.OnMsg(meta.DeleteChunk, meta.MsgCallback(func(args ...interface{}) error {
		chunkid := args[0].(uint64)
		length := args[1].(uint32)
		return store.Remove(chunkid, int(length))
	}))

	// Create virtual filesystem with mountpoint for S3 and chunk size
	conf := &vfs.Config{
		Meta: &meta.Config{
			IORetries: 10,
		},
		Format:     format,
		Mountpoint: mp,
		Primary: &vfs.StorageConfig{
			Name:      format.Storage,
			Endpoint:  format.Bucket,
			AccessKey: format.AccessKey,
			SecretKey: format.AccessKey,
		},
		Chunk: &chunkConf,
	}
	vfs.Init(conf, m, store)

	installHandler(mp)
	// Attach FUSE to VFS, this is what initializes our Spinel engine core with FUSE
	err = fuse.Main(conf, c.String("o"), c.Float64("attrcacheto"), c.Float64("entrycacheto"), c.Float64("direntrycacheto"))
	if err != nil {
		logger.Errorf("%s", err)
		os.Exit(1)
	}
	return nil
}

func mountFlags() *cli.Command {
	/*
		Mountpoint flags to mount a file system
	*/
	var defaultCacheDir = "/var/spinelCache"
	if runtime.GOOS == "darwin" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			logger.Fatalf("%v", err)
			return nil
		}
		defaultCacheDir = path.Join(homeDir, ".spinel", "cache")
	}
	return &cli.Command{
		Name:      "mount",
		Usage:     "mount a volume",
		ArgsUsage: "REDIS-URL MOUNTPOINT",
		Action:    mount,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "d",
				Usage: "run in background",
			},
			&cli.StringFlag{
				Name:  "o",
				Usage: "other fuse options",
			},
			&cli.Float64Flag{
				Name:  "attrcacheto",
				Value: 1.0,
				Usage: "attributes cache timeout in seconds",
			},
			&cli.Float64Flag{
				Name:  "entrycacheto",
				Value: 1.0,
				Usage: "file entry cache timeout in seconds",
			},
			&cli.Float64Flag{
				Name:  "direntrycacheto",
				Value: 1.0,
				Usage: "dir entry cache timeout in seconds",
			},

			&cli.IntFlag{
				Name:  "getTimeout",
				Value: 60,
				Usage: "the max number of seconds to download an object",
			},
			&cli.IntFlag{
				Name:  "putTimeout",
				Value: 60,
				Usage: "the max number of seconds to upload an object",
			},
			&cli.IntFlag{
				Name:  "ioretries",
				Value: 30,
				Usage: "number of retries after network failure",
			},
			&cli.IntFlag{
				Name:  "maxUpload",
				Value: 20,
				Usage: "number of connections to upload",
			},
			&cli.IntFlag{
				Name:  "bufferSize",
				Value: 300,
				Usage: "total read/write buffering in MB",
			},
			&cli.IntFlag{
				Name:  "prefetch",
				Value: 3,
				Usage: "prefetch N blocks in parallel",
			},

			&cli.BoolFlag{
				Name:  "writeback",
				Usage: "Upload objects in background",
			},
			&cli.StringFlag{
				Name:  "cacheDir",
				Value: defaultCacheDir,
				Usage: "directory to cache object",
			},
			&cli.IntFlag{
				Name:  "cacheSize",
				Value: 1 << 10,
				Usage: "size of cached objects in MiB",
			},
			&cli.Float64Flag{
				Name:  "freeSpace",
				Value: 0.1,
				Usage: "min free space (ratio)",
			},
			&cli.BoolFlag{
				Name:  "partialOnly",
				Usage: "cache only random/small read",
			},
		},
	}
}
