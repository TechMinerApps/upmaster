package main

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/TechMinerApps/upmaster/modules/database"
	"github.com/TechMinerApps/upmaster/modules/log"
	"github.com/TechMinerApps/upmaster/modules/utils"
	"github.com/TechMinerApps/upmaster/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// UpMaster is a object of entire app
type UpMaster struct {
	Config Config

	DB          *gorm.DB
	InfluxDB    database.InfluxDB
	HTTPServer  *http.Server
	HTTPHandler *gin.Engine

	viper  *viper.Viper
	logger log.Logger
	wg     *sync.WaitGroup
}

// Config has all the configuration required by UpMaster
type Config struct {
	Port            int
	RDBMSConfig     database.RDBMSConfig    `mapstructure:"rdbms"`
	InfluxDBConfig  database.InfluxDBConfig `mapstructure:"influxdb"`
	OAuthGCInterval int                     `mapstructure:"oauth_interval"`
}

// NewUpMaster is used to generate a UpMaster object
// no error is returned, so error must be handled within NewUpMaster
func NewUpMaster() *UpMaster {
	var app UpMaster

	app.setupLogger()
	app.setupViper()
	app.setupDB()
	app.setupRouter()

	app.wg = &sync.WaitGroup{}

	return &app
}

// Start starts the instance of UpMaster non-blocking
// waitgroup inside UpMaster is added by 1
func (u *UpMaster) Start() {

	u.HTTPServer = &http.Server{
		Addr:    ":" + strconv.Itoa(u.Config.Port),
		Handler: u.HTTPHandler,
	}

	u.wg.Add(1)

	go func() {
		if err := u.HTTPServer.ListenAndServe(); err != http.ErrServerClosed && err != nil {
			u.logger.Fatalf("HTTP Server Listen Error: %v", err)
		}
	}()
	u.logger.Infof("UpMaster Started")

	return
}

// Stop does the graceful shutdown
// waitgroup done here should reduce the waitgroup to 0
func (u *UpMaster) Stop(sig os.Signal) {

	// Log the reason for stopping
	u.logger.Infof("Recieved signal %s", sig.String())

	// Graceful shutdown http server, with a  5 seconds timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := u.HTTPServer.Shutdown(ctx); err != nil {
		u.logger.Errorf("Server Closed With Error: %s", err.Error())
	}
	u.logger.Infof("UpMaster Shutdown")
	u.wg.Done()
}

func (u *UpMaster) setupRouter() {

	// Parse Config
	var routerCfg router.Config
	routerCfg.DB = u.DB
	routerCfg.DBName = u.Config.RDBMSConfig.MySQLConfig.DBName
	routerCfg.OAuthGCInterval = u.Config.OAuthGCInterval

	// Create new handler
	var err error
	u.HTTPHandler, err = router.NewRouter(routerCfg)
	if err != nil {
		u.logger.Fatalf("Router Config Failed: %v", err)
	}
}

func (u *UpMaster) setupDB() {
	var err error
	u.DB, err = database.NewRDBMSConnection(u.Config.RDBMSConfig)
	if err != nil {
		u.logger.Fatalf("Unable to establish RDBMS connection: %v", err)
	}
	u.InfluxDB, err = database.NewInfluxDBConnection(u.Config.InfluxDBConfig)
	if err != nil {
		u.logger.Fatalf("Unable to establish InfluxDB connection: %v", err)
	}
}

func (u *UpMaster) setupViper() {
	u.viper = viper.New()
	pflag.String("config", "config", "config file name")
	pflag.Parse()
	u.viper.BindPFlags(pflag.CommandLine)

	if u.viper.IsSet("config") {
		u.viper.SetConfigFile(u.viper.GetString("config"))
	} else {
		u.viper.SetConfigName("config")
		u.viper.SetConfigType("yaml")
		u.viper.AddConfigPath(utils.AbsPath(""))
		u.viper.AddConfigPath("/etc/upmaster")
	}

	u.viper.SetEnvPrefix("UPMASTER")
	u.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	u.viper.AutomaticEnv()

	if err := u.viper.ReadInConfig(); err != nil {
		// Used logger here, so setupLogger before setupViper
		u.logger.Fatalf("Unable to read in config: %v", err)
	}

	if err := u.viper.Unmarshal(&u.Config); err != nil {
		u.logger.Fatalf("Unable to decode into struct: %v", err)
	}
}

func (u *UpMaster) setupLogger() {
	u.logger = log.NewLogger()
}
