// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cron

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	log "unknwon.dev/clog/v2"

	"github.com/wuhan005/oblivion/internal/db"
)

func Start(k8sClient *kubernetes.Clientset) {
	go start(k8sClient)
}

func start(k8sClient *kubernetes.Clientset) {
	ctx := context.Background()

	for {
		pods, err := db.Pods.GetExpired(ctx)
		if err != nil {
			log.Error("Failed to get expired pods: %v", err)
			continue
		}

		for _, pod := range pods {
			namespace := fmt.Sprintf("%s-%s", pod.Image.UID, pod.User.Domain)

			// Delete pods in cluster.
			if err := k8sClient.CoreV1().Pods(namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{}); err != nil {
				log.Error("Failed to automatically delete pod in cluster: %v", err)
			}

			if err := db.Pods.Delete(ctx, pod.ID); err != nil {
				log.Error("Failed to automatically delete pod: %v", err)
			}
		}
		time.Sleep(5 * time.Second)
	}
}
