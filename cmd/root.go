package cmd

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/go-redis/redis/v7"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"github.com/asherda/lightwalletd/common"
	"github.com/asherda/lightwalletd/common/logging"
	"github.com/asherda/lightwalletd/frontend"
	"github.com/asherda/lightwalletd/walletrpc"
)

var cfgFile string
var logger = logrus.New()

// StartOpts contains the command line options set for this run of lightwalletd
var StartOpts *common.Options

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "lightwalletd",
	Short: "Lightwalletd is a backend service to the VerusCoin blockchain",
	Long: `Lightwalletd is a backend service that provides a 
         bandwidth-efficient interface to the VerusCoin blockchain`,
	Run: func(cmd *cobra.Command, args []string) {
		StartOpts = &common.Options{
			BindAddr:          viper.GetString("bind-addr"),
			TLSCertPath:       viper.GetString("tls-cert"),
			TLSKeyPath:        viper.GetString("tls-key"),
			LogLevel:          viper.GetUint64("log-level"),
			LogFile:           viper.GetString("log-file"),
			NoTLSVeryInsecure: viper.GetBool("no-tls-very-insecure"),
			VerusdURL:         viper.GetString("verusd-url"),
			ChainName:         viper.GetString("chain-name"),
			VerusdUser:        viper.GetString("verusd-user"),
			VerusdPassword:    viper.GetString("verusd-password"),
			RedisURL:          viper.GetString("redis-url"),
			RedisPassword:     viper.GetString("redis-password"),
			RedisDB:           viper.GetInt("redis-db"),
			CacheSize:         viper.GetInt("cache-size"),
			VerusConfPath:     viper.GetString("verus-conf-path"),
		}

		common.Log.Debugf("Options: %#v\n", StartOpts)
		if len(StartOpts.VerusdURL) < 1 && len(StartOpts.RedisURL) < 1 {
			os.Stderr.WriteString(fmt.Sprintf("Configuration failure: At least one of --verusd-url or --redis-url command line options must be specified. Both can be specified, but not neither."))
			os.Exit(1)
		}

		var filesThatShouldExist []string
		if StartOpts.NoVerusd {
			filesThatShouldExist = []string{
				StartOpts.LogFile,
			}
		} else {
			filesThatShouldExist = []string{
				StartOpts.TLSCertPath,
				StartOpts.TLSKeyPath,
				StartOpts.LogFile,
			}
		}

		for _, filename := range filesThatShouldExist {
			if !fileExists(StartOpts.LogFile) {
				os.OpenFile(StartOpts.LogFile, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
			}
			if StartOpts.NoTLSVeryInsecure && (filename == StartOpts.TLSCertPath || filename == StartOpts.TLSKeyPath) {
				continue
			}
			if !fileExists(filename) {
				os.Stderr.WriteString(fmt.Sprintf("\n  ** File does not exist: %s\n\n", filename))
				os.Exit(1)
			}
		}

		// Start server and block, or exit
		if err := startServer(StartOpts); err != nil {
			common.Log.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("couldn't create server")
		}
	},
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func startServer(startOpts *common.Options) error {
	if startOpts.LogFile != "" {
		// instead write parsable logs for logstash/splunk/etc
		output, err := os.OpenFile(startOpts.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			common.Log.WithFields(logrus.Fields{
				"error": err,
				"path":  startOpts.LogFile,
			}).Fatal("couldn't open log file")
		}
		defer output.Close()
		logger.SetOutput(output)
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	logger.SetLevel(logrus.Level(startOpts.LogLevel))
	// gRPC initialization
	var server *grpc.Server

	if startOpts.NoTLSVeryInsecure {
		common.Log.Warningln("Starting insecure server")
		fmt.Println("Starting insecure server")
		server = grpc.NewServer(logging.LoggingInterceptor())
	} else {
		transportCreds, err := credentials.NewServerTLSFromFile(startOpts.TLSCertPath, startOpts.TLSKeyPath)
		if err != nil {
			common.Log.WithFields(logrus.Fields{
				"cert_file": startOpts.TLSCertPath,
				"key_path":  startOpts.TLSKeyPath,
				"error":     err,
			}).Fatal("couldn't load TLS credentials")
		}
		server = grpc.NewServer(grpc.Creds(transportCreds), logging.LoggingInterceptor())
	}

	// Enable reflection for debugging
	if startOpts.LogLevel >= uint64(logrus.WarnLevel) {
		reflection.Register(server)
	}

	var cache *common.BlockCache

	redisOpts := &redis.Options{
		Addr:     startOpts.RedisURL,
		Password: startOpts.RedisPassword,
		DB:       startOpts.RedisDB,
	}

	var rpcClient *rpcclient.Client
	if !startOpts.NoVerusd {
		// Initialize verusd RPC client. Right now (April 2020) this is only for
		// sending transactions and handling identity, but in the future it could
		// back a different type of block streamer.
		var err error
		rpcClient, err = frontend.NewVRPCFromConf(startOpts.ChainName, startOpts.VerusdURL, startOpts.VerusdUser, startOpts.VerusdPassword)
		if err != nil {
			common.Log.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("setting up RPC connection to verusd")
		}

		// Get the sapling activation height from the RPC
		// (this first RPC also verifies that we can communicate with verusd)
		saplingHeight, blockHeight, subChainName, branchID := common.GetSaplingInfo(rpcClient)
		common.Log.Info("Got sapling height from verusd ", saplingHeight, ", chainName ", startOpts.ChainName, ", subchain ", subChainName, " branchID ", branchID)

		cache, err := common.NewBlockCache(startOpts.ChainName, startOpts.CacheSize, saplingHeight, blockHeight, rpcClient, redisOpts)
		cachedBlockHeight := common.UpdateRedisValues(cache.RedisClient, saplingHeight, blockHeight, startOpts.ChainName, subChainName, branchID)
		var cacheStart int

		// TODO: start earlier when cachedBlockHeight is set (for ingesting) to check for reorgs.
		//
		// If we have nothing cahced then start from saplingHeight
		// Otherwise start from where the cache got to
		if cachedBlockHeight > 1 {
			cacheStart = cachedBlockHeight
		} else {
			cacheStart = 1
		}
		// Start the block cache importer (ingestor or BlockIngestor) at the highest cached block
		go common.BlockIngestor(cache, cacheStart, 0 /*loop forever*/)
	} else {
		if len(StartOpts.RedisURL) > 0 {
			redisClient, err := common.GetCheckedRedisClient(redisOpts)
			if err != nil {
				os.Stderr.WriteString(fmt.Sprintf("\n  ** redis is enabled but lightwalletd is unable to connect to the redis host\n\n"))
				os.Exit(1)
			}

			saplingHeight := common.CheckRedisIntResult(redisClient, StartOpts.ChainName+"-saplingHeight")
			blockHeight := common.CheckRedisIntResult(redisClient, StartOpts.ChainName+"-blockHeight")
			branchID := common.CheckRedisStringResult(redisClient, StartOpts.ChainName+"-branchID")
			subChainName := common.CheckRedisStringResult(redisClient, StartOpts.ChainName+"-branchID")

			common.Log.Info("Got sapling height from redis ", saplingHeight, ", blockHeight ", blockHeight, ", chainName \"", startOpts.ChainName, "\", subChain ", subChainName, "branchID ", branchID)
			cache, err = common.NewBlockCache(startOpts.ChainName, startOpts.CacheSize, saplingHeight, blockHeight, rpcClient, redisOpts)
			if err != nil {
				os.Stderr.WriteString(fmt.Sprintf("\n  ** unable to create cache for " + startOpts.ChainName + "\n\n"))
				os.Exit(1)
			}
		}
	}

	// Compact transaction service initialization
	service, err := frontend.NewLwdStreamer(cache)
	if err != nil {
		common.Log.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("couldn't create backend")
	}

	// Register service
	walletrpc.RegisterCompactTxStreamerServer(server, service)

	// Start listening
	listener, err := net.Listen("tcp", startOpts.BindAddr)
	if err != nil {
		common.Log.WithFields(logrus.Fields{
			"bind_addr": startOpts.BindAddr,
			"error":     err,
		}).Fatal("couldn't create listener")
	}

	// Signal handler for graceful stops
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-signals
		common.Log.WithFields(logrus.Fields{
			"signal": s.String(),
		}).Info("caught signal, stopping gRPC server")
		os.Exit(1)
	}()

	common.Log.WithFields(logrus.Fields{
		"gitCommit": common.GitCommit,
		"buildDate": common.BuildDate,
		"buildUser": common.BuildUser,
	}).Infof("Starting gRPC server version %s on %s", common.Version, startOpts.BindAddr)

	err = server.Serve(listener)
	if err != nil {
		common.Log.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("gRPC server exited")
	}
	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is current directory, lightwalletd.yaml)")
	rootCmd.Flags().String("bind-addr", "127.0.0.1:18232", "the address to listen on")
	rootCmd.Flags().String("tls-cert", "./cert.pem", "the path to a TLS certificate")
	rootCmd.Flags().String("tls-key", "./cert.key", "the path to a TLS key file")
	rootCmd.Flags().Int("log-level", int(logrus.InfoLevel), "log level (logrus 1-7)")
	rootCmd.Flags().String("log-file", "./server.log", "log file to write to")
	rootCmd.Flags().Bool("no-tls-very-insecure", false, "run without the required TLS certificate, only for debugging, DO NOT use in production")
	rootCmd.Flags().Bool("no-verusd", false, "run without using verusd, getting data from redis")
	rootCmd.Flags().String("verus-conf-path", "~/.komodo/VRSC/VRSC.conf", "path + file name for VRSC.conf file")
	rootCmd.Flags().String("verusd-url", "127.0.0.1:2786", "the verusd RPC address and port to connect to")
	rootCmd.Flags().String("chain-name", "VRSC", "chain name, defaults to VRSC")
	rootCmd.Flags().String("verusd-user", "VERUSDUSER", "verusd user access credential")
	rootCmd.Flags().String("verusd-password", "VERUSDPASSWORD", "verusd password access credential")
	rootCmd.Flags().String("redis-url", "", "URL of redis server including port; leave out or set to \"\" to disable redis")
	rootCmd.Flags().String("redis-password", "", "password for redis server if needed")
	rootCmd.Flags().Int("redis-db", 0, "DB number for redis cache")
	rootCmd.Flags().Int("cache-size", 15000000, "number of blocks to hold in the cache")

	viper.BindPFlag("bind-addr", rootCmd.Flags().Lookup("bind-addr"))
	viper.SetDefault("bind-addr", "127.0.0.1:9067")
	viper.BindPFlag("tls-cert", rootCmd.Flags().Lookup("tls-cert"))
	viper.SetDefault("tls-cert", "./cert.pem")
	viper.BindPFlag("tls-key", rootCmd.Flags().Lookup("tls-key"))
	viper.SetDefault("tls-key", "./cert.key")
	viper.BindPFlag("log-level", rootCmd.Flags().Lookup("log-level"))
	viper.SetDefault("log-level", int(logrus.InfoLevel))
	viper.BindPFlag("log-file", rootCmd.Flags().Lookup("log-file"))
	viper.SetDefault("log-file", "./server.log")
	viper.BindPFlag("verus-conf-path", rootCmd.Flags().Lookup("verus-conf-path"))
	viper.SetDefault("verus-conf-path", "./VRSC.conf")
	viper.BindPFlag("zcash-conf-path", rootCmd.Flags().Lookup("zcash-conf-path"))
	viper.SetDefault("zcash-conf-path", "./zcash.conf")
	viper.BindPFlag("no-tls-very-insecure", rootCmd.Flags().Lookup("no-tls-very-insecure"))
	viper.SetDefault("no-tls-very-insecure", false)
	viper.BindPFlag("verusd-url", rootCmd.Flags().Lookup("verusd-url"))
	viper.SetDefault("verusd-url", "127.0.0.1:2786")
	viper.BindPFlag("chain-name", rootCmd.Flags().Lookup("chain-name"))
	viper.SetDefault("chain-name", "VRSC")
	viper.BindPFlag("verusd-user", rootCmd.Flags().Lookup("verusd-user"))
	viper.SetDefault("verusd-user", "VERUSDUSER")
	viper.BindPFlag("verusd-password", rootCmd.Flags().Lookup("verusd-password"))
	viper.SetDefault("verusd-password", "VERUSDPASSWORD")
	viper.BindPFlag("no-verusd", rootCmd.Flags().Lookup("no-verusd"))
	viper.SetDefault("no-verusd", false)
	viper.BindPFlag("redis-url", rootCmd.Flags().Lookup("redis-url"))
	viper.SetDefault("redis-url", "")
	viper.BindPFlag("redis-password", rootCmd.Flags().Lookup("redis-password"))
	viper.SetDefault("redis-password", "")
	viper.BindPFlag("redis-db", rootCmd.Flags().Lookup("redis-db"))
	viper.SetDefault("redis-db", 0)
	viper.BindPFlag("cache-size", rootCmd.Flags().Lookup("cache-size"))
	viper.SetDefault("cache-size", 15000000)

	logger.SetFormatter(&logrus.TextFormatter{
		//DisableColors:          true,
		FullTimestamp:          true,
		DisableLevelTruncation: true,
	})

	onexit := func() {
		fmt.Printf("Lightwalletd died with a Fatal error. Check logfile for details.\n")
	}

	common.Log = logger.WithFields(logrus.Fields{
		"app": "lightwalletd",
	})

	logrus.RegisterExitHandler(onexit)

	// Indirect function for test mocking (so unit tests can talk to stub functions)
	common.Sleep = time.Sleep
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Look in the current directory for a configuration file
		viper.AddConfigPath(".")
		// Viper auto appends extention to this config name
		// For example, lightwalletd.yml
		viper.SetConfigName("lightwalletd")
	}

	// Replace `-` in config options with `_` for ENV keys
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv() // read in environment variables that match
	// If a config file is found, read it in.
	var err error
	if err = viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
