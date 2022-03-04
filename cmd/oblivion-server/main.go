package main

import (
	"io/ioutil"
	"net"
	"os"

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

	const (
		tokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	)
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	token, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		log.Fatal("Failed to get token: %v", err)
	}

	k8sClient, err := kubernetes.NewForConfig(&rest.Config{
		Host:            "http://" + net.JoinHostPort(host, port),
		TLSClientConfig: rest.TLSClientConfig{},
		BearerToken:     string(token),
		BearerTokenFile: tokenFile,
	})
	if err != nil {
		log.Fatal("Failed to get k8s client: %v", err)
	}
	database, err := db.Init()
	if err != nil {
		log.Fatal("Failed to init database: %v", err)
	}

	f := flamego.Classic()
	f.Use(flamego.Renderer())
	f.Map(k8sClient)

	f.Use(context.Contexter(database))

	f.Get("/health", func() {})
	f.Group("/api", func() {
		f.Group("/env/{uid}", func() {
			f.Combo("").
				Get(route.CreatePod).
				Delete(route.DeletePod)
		}, route.Enver)
	}, route.UserAuther)

	f.Run(4000)
}
