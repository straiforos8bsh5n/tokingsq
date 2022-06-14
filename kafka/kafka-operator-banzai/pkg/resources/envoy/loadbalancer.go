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

package envoy

import (
	"fmt"

	"github.com/banzaicloud/kafka-operator/api/v1beta1"
	"github.com/banzaicloud/kafka-operator/pkg/resources/templates"
	envoyutils "github.com/banzaicloud/kafka-operator/pkg/util/envoy"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	corev1 "k8s.io/api/core/v1"
)

// loadBalancer return a Loadbalancer service for Envoy
func (r *Reconciler) loadBalancer(log logr.Logger) runtime.Object {

	exposedPorts := getExposedServicePorts(r.KafkaCluster.Spec.ListenersConfig.ExternalListeners, r.KafkaCluster.Spec.Brokers)

	service := &corev1.Service{
		ObjectMeta: templates.ObjectMetaWithAnnotations(envoyutils.EnvoyServiceName, map[string]string{}, r.KafkaCluster.Spec.EnvoyConfig.GetAnnotations(), r.KafkaCluster),
		Spec: corev1.ServiceSpec{
			Selector:                 map[string]string{"app": "envoy"},
			Type:                     corev1.ServiceTypeLoadBalancer,
			Ports:                    exposedPorts,
			LoadBalancerSourceRanges: r.KafkaCluster.Spec.EnvoyConfig.GetLoadBalancerSourceRanges(),
		},
	}
	return service
}

func getExposedServicePorts(extListeners []v1beta1.ExternalListenerConfig, brokers []v1beta1.Broker) []corev1.ServicePort {
	var exposedPorts []corev1.ServicePort

	for _, eListener := range extListeners {
		for _, broker := range brokers {
			exposedPorts = append(exposedPorts, corev1.ServicePort{
				Name:       fmt.Sprintf("broker-%d", broker.Id),
				Port:       eListener.ExternalStartingPort + broker.Id,
				TargetPort: intstr.FromInt(int(eListener.ExternalStartingPort + broker.Id)),
				Protocol:   corev1.ProtocolTCP,
			})
		}
	}
	return exposedPorts
}
