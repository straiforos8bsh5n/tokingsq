// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"net"
	"net/http"

	"github.com/banzaicloud/kafka-operator/internal/alertmanager"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	receiverAddr = ":9001"
)

// AController implements Runnable
type AController struct {
	Client client.Client
}

// SetAlertManagerWithManager creates a new Alertmanager Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func SetAlertManagerWithManager(mgr manager.Manager) error {
	return mgr.Add(AController{Client: mgr.GetClient()})
}

// Start initiates the alertmanager controller
func (c AController) Start(<-chan struct{}) error {
	logf.SetLogger(logf.ZapLogger(false))
	log := logf.Log.WithName("alertmanager-entrypoint")

	ln, _ := net.Listen("tcp", receiverAddr)
	httpServer := &http.Server{Handler: alertmanager.NewApp(log, c.Client)}
	return httpServer.Serve(ln)
}
