package main

import (
	"github.com/flamego/flamego"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	log "unknwon.dev/clog/v2"

	"github.com/wuhan005/oblivion/internal/context"
	"github.com/wuhan005/oblivion/internal/db"
	"github.com/wuhan005/oblivion/internal/route"
)

func main() {
	defer log.Stop()
	err := log.NewConsole()
	if err != nil {
		panic(err)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal("Failed to get in cluster config: %v", err)
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal("Failed to get k8s client: %v", err)
	}
	database, err := db.Init()
	if err != nil {
		log.Fatal("Failed to init database: %v", err)
	}

	f := flamego.Classic()
	f.Map(k8sClient)

	f.Use(context.Contexter(database))

	f.Group("/api", func() {
		f.Group("/env/{uid}", func() {
			f.Combo("").
				Get(route.GetPod).
				Post(route.CreatePod).
				Delete(route.DeletePod)
		}, route.Enver)
	}, route.UserAuther)

	f.Run(4000)
}
