package main

import (
    "app/flow"
    "conf"
    "io/ioutil"
    "os"
    "time"
    "util"
)

var logger = util.NewLogger("main")

func main() {
    confPath := readEnv("CONFIG", "./conf/dev.json")

    if byts, err := ioutil.ReadFile(confPath); err == nil {
        config := conf.NewConf(byts)
        feeds := config.GetList("feed_list")
        entry, _ := flow.NewBaseFlow()

        for {
            for _, feed := range feeds {
                entry <- feed
            }

            time.Sleep(time.Hour)
        }
    } else {
        logger.Fatalf("error with reading config file[%s]: %s", confPath, err)
    }
}

func readEnv(name string, def string) (ret string) {
    val, exist := os.LookupEnv(name)

    if exist {
        ret = val

        logger.Logf("ENV [%s] read as [%s]", name, val)
    } else {
        ret = def

        logger.Logf("ENV [%s] not found, use [%s]", name, def)
    }

    return
}