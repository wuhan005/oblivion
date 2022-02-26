// Copyright 2022 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package route

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	log "unknwon.dev/clog/v2"

	"github.com/wuhan005/oblivion/internal/context"
	"github.com/wuhan005/oblivion/internal/db"
)

func UserAuther(ctx context.Context) error {
	token := ctx.Query("token")
	user, err := db.Users.GetByToken(ctx.Request().Context(), token)
	if err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			return ctx.Error(40400, "token is invalid")
		}
		log.Error("Failed to get user by token: %v", err)
		return ctx.ServerError()
	}

	ctx.Map(user)
	return nil
}

func Enver(ctx context.Context) error {
	imageUID := ctx.Param("uid")
	image, err := db.Images.GetByUID(ctx.Request().Context(), imageUID)
	if err != nil {
		if errors.Is(err, db.ErrImageNotFound) {
			return ctx.Error(40400, "Environment not found")
		}
		log.Error("Failed to get image by uid: %v", err)
		return ctx.ServerError()
	}
	ctx.Map(image)
	return nil
}

func GetPod(ctx context.Context, user *db.User, image *db.Image) error {
	pods, err := db.Pods.Get(ctx.Request().Context(), db.GetPodsOptions{
		UserID:  user.ID,
		ImageID: image.ID,
	})
	if err != nil {
		log.Error("Failed to get pods: %v", err)
		return ctx.ServerError()
	}

	if len(pods) == 0 {
		return ctx.Error(40400, "Pod not found")
	}
	return ctx.Success(pods[0])
}

func CreatePod(ctx context.Context, user *db.User, image *db.Image, k8sClient *kubernetes.Clientset) error {
	pods, err := db.Pods.Get(ctx.Request().Context(), db.GetPodsOptions{
		UserID:  user.ID,
		ImageID: image.ID,
	})
	if err != nil {
		log.Error("Failed to get pods: %v", err)
		return ctx.ServerError()
	}
	if len(pods) != 0 {
		return ctx.Error(40300, "Pod has been created")
	}

	// Create pod in cluster.
	namespace := image.Namespace
	podName := fmt.Sprintf("gamebox-%s-%s", namespace, user.Domain)
	podPort := image.Port
	_, err = k8sClient.CoreV1().Pods(namespace).Create(ctx.Request().Context(), &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels: map[string]string{
				"team_token": user.Token,
				"image_uid":  image.UID,
				"image_name": image.Name,
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  podName,
					Image: image.Name,
					Ports: []v1.ContainerPort{
						{
							ContainerPort: podPort,
						},
					},
				},
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		log.Error("Failed to create pod in cluster: %v", err)
		return ctx.ServerError()
	}

	address := user.Domain + "." + image.Domain

	// Create ingress for pod with address domain.
	_, err = k8sClient.NetworkingV1beta1().Ingresses(namespace).Create(ctx.Request().Context(), &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels: map[string]string{
				"team_token": user.Token,
				"image_uid":  image.UID,
				"image_name": image.Name,
			},
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: address,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: v1beta1.IngressBackend{
										ServiceName: podName,
										ServicePort: intstr.FromInt(int(podPort)),
									},
								},
							},
						},
					},
				},
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		log.Error("Failed to create pod ingress: %v", err)
		return ctx.ServerError()
	}

	pod, err := db.Pods.Create(ctx.Request().Context(), db.CreatePodOptions{
		UserID:    user.ID,
		ImageID:   image.ID,
		Name:      podName,
		Address:   address,
		ExpiredAt: time.Now().Add(1 * time.Hour),
	})
	if err != nil {
		log.Error("Failed to create pod: %v", err)
		return ctx.ServerError()
	}
	return ctx.Success(pod)
}

func DeletePod(ctx context.Context, user *db.User, image *db.Image, k8sClient *kubernetes.Clientset) error {
	pods, err := db.Pods.Get(ctx.Request().Context(), db.GetPodsOptions{
		UserID:  user.ID,
		ImageID: image.ID,
	})
	if err != nil {
		log.Error("Failed to get pods: %v", err)
		return ctx.ServerError()
	}

	if len(pods) == 0 {
		return ctx.Error(40400, "Pod not found")
	}
	pod := pods[0]

	// Delete pods in cluster.
	namespace := image.Namespace
	if err := k8sClient.CoreV1().Pods(namespace).Delete(ctx.Request().Context(), pod.Name, metav1.DeleteOptions{}); err != nil {
		log.Error("Failed to delete pod in cluster: %v", err)
	}

	if err := db.Pods.Delete(ctx.Request().Context(), pod.ID); err != nil {
		log.Error("Failed to delete pod: %v", err)
		return ctx.ServerError()
	}
	return ctx.Success()
}
