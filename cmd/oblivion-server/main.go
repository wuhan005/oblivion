package main

import (
	"github.com/flamego/flamego"
	"github.com/flamego/session"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	log "unknwon.dev/clog/v2"

	"github.com/wuhan005/oblivion/internal/context"
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

	f := flamego.Classic()
	f.Map(k8sClient)

	sessioner := session.Sessioner()
	f.Use(
		sessioner,
		context.Contexter(nil),
	)

	f.Get("/{id}")
	f.Group("/api", func() {
		f.Combo("/{id}").
			Get().
			Post().
			Delete()
	})

	f.Combo("/sign-in").Get().Post()
	f.Group("/admin", func() {
		f.Get("/")

		f.Group("/user", func() {
			f.Combo("/").Get().Post()
			f.Post("/delete")
			f.Post("/batch")
		})

		f.Group("/pod", func() {
			f.Combo("/").Get().Post()
			f.Post("/delete")
		})

		f.Group("/image", func() {
			f.Combo("/").Get(route.ListImages).Post()
		})
	})

	f.Run(4000)
}
