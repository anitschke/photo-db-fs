package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/anitschke/photo-db-fs/db"
	_ "github.com/anitschke/photo-db-fs/db/register"
	"github.com/anitschke/photo-db-fs/photofs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var mountPointFlag = flag.String("mount-point", "", "location where photo-db-fs file system will be mounted")
var dbTypeFlag = flag.String("db-type", "", "type of photo database photo-db-fs will use for querying")
var dbSource = flag.String("db-source", "", "source of the database photo-db-fs will use for querying, ie for local databases this is the path to the database")

var logLevelFlag = flag.String("log-level", "", "debugging logging level")

func main() {
	flag.Parse()

	if *mountPointFlag == "" {
		fmt.Println("--mount-point flag is required")
		return
	}
	if *dbTypeFlag == "" {
		fmt.Println("--db-type flag is required")
		return
	}
	if *dbSource == "" {
		fmt.Println("--db-source flag is required")
		return
	}

	logger, err := setupLogging()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer logger.Sync()

	ctx := context.Background()

	db, err := db.New(*dbTypeFlag, *dbSource)
	if err != nil {
		zap.L().Fatal("failed to connect to database", zap.Error(err))
	}
	defer func() {
		err := db.Close()
		if err != nil {
			zap.L().Error("error closing database", zap.Error(err))
		}
	}()

	server, err := photofs.Mount(ctx, *mountPointFlag, db)
	if err != nil {
		zap.L().Fatal("failed to mount file system", zap.Error(err))
		return
	}
	zap.L().Debug("photo-db-fs mounted", zap.String("mountPoint", *mountPointFlag))

	doneC := make(chan struct{})
	killC := make(chan os.Signal, 10) // As per signal package doc must be a big enough buffer

	// Listen for ctrl+c and other signals to kill process and shutdown and
	// unmount if we hear these. But we also have a call to server.Wait() below.
	// So if that returns it means the server was unmounted and we don't need to
	// unmount ourselves.
	signal.Notify(killC, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP)

	go func() {
		// Sometimes killing will fail if someone is still hanging on to an
		// inode of a file in our fuse file system. If this is the case we will
		// just wait for more requests to kill and try again when we get another
		// one.
		for {
			select {
			case <-doneC:
				return
			case sig := <-killC:
				zap.L().Debug("caught signal", zap.Stringer("signal", sig))
				err := server.Unmount()
				if err != nil {
					zap.L().Error(
						"Failed to unmount. This is likely because another program "+
							"is still hanging on to a file handle / inode to a file "+
							"served by this fuse server. Please release that file "+
							"handle / close that program and then try again.",
						zap.Error(err))
				} else {
					return
				}
			}
		}
	}()

	// Wait until unmount before exiting
	server.Wait()
	close(doneC)
}

func setupLogging() (*zap.Logger, error) {
	zapConfig := zap.NewProductionConfig()

	level := zap.ErrorLevel
	if *logLevelFlag != "" {
		if err := level.UnmarshalText([]byte(*logLevelFlag)); err != nil {
			return nil, fmt.Errorf("failed to parse log-level: %w", err)
		}

	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	encoderConfig := zap.NewProductionEncoderConfig()

	// I find that the caller and stacktrace adds a lot of extra noise and isn't
	// super useful for such a small code base.
	encoderConfig.CallerKey = zapcore.OmitKey
	encoderConfig.StacktraceKey = zapcore.OmitKey

	// Better format for time
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zapConfig.EncoderConfig = encoderConfig

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}
	zap.ReplaceGlobals(logger)
	return logger, nil
}
