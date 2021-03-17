package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/viper"

	"demo.hello/apps/viper/configs"
)

func demo01() {
	// Set the path to look for the configurations file
	viper.AddConfigPath("configs")
	// Set the file name of the configurations file
	viper.SetConfigName("config")

	// Enable viper to read Environment Variables
	viper.AutomaticEnv()

	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic("Config file not found")
		} else {
			panic(fmt.Errorf("Error reading config file, %v", err))
		}
	}

	// Set undefined variables
	viper.SetDefault("database.dbname", "test_db")

	// EXAMPLE_PATH or EXAMPLE_VAR will be override as (export EXAMPLE_VAR="viper_test")
	fmt.Println("Reading variables without using the model..")
	fmt.Println("Database is\t", viper.GetString("database.dbname"))
	fmt.Println("Port is\t\t", viper.GetInt("server.port"))
	fmt.Println("EXAMPLE_PATH is\t", viper.GetString("EXAMPLE_PATH"))
	fmt.Println("EXAMPLE_VAR is\t", viper.GetString("EXAMPLE_VAR"))

	var configuration configs.Configurations
	if err := viper.Unmarshal(&configuration); err != nil {
		panic(fmt.Sprintf("Unable to decode into struct, %v\n", err))
	}

	fmt.Println("\nReading variables using the model..")
	fmt.Println("Database is\t", configuration.Database.DBName)
	fmt.Println("Port is\t\t", configuration.Server.Port)
	fmt.Println("EXAMPLE_PATH is\t", configuration.ExamplePath)
	fmt.Println("EXAMPLE_VAR is\t", configuration.ExampleVar)
}

func demo02() {
	path := "configs/config.yml"
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("Read config file fail: %s\n", err.Error()))
	}

	viper.SetConfigType("yml")

	content := os.ExpandEnv(string(b))
	if err := viper.ReadConfig(strings.NewReader(content)); err != nil {
		panic(fmt.Sprintf("Parse config file fail: %s\n", err.Error()))
	}

	fmt.Println("env.user=" + viper.GetString("env.user"))
	fmt.Println("env.home=" + viper.GetString("env.home"))
}

func main() {
	demo02()
	fmt.Println("viper done")
}
