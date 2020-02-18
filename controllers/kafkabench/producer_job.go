/*
Copyright 2019 The xridge kubestone contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kafkabench

import (
	"fmt"
	perfv1alpha1 "github.com/xridge/kubestone/api/v1alpha1"
	"github.com/xridge/kubestone/pkg/k8s"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func NewProducerJob(cr *perfv1alpha1.KafkaBench, ts *perfv1alpha1.KafkaTestSpec) *batchv1.Job {
	objectMeta := metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s-%s-producer", cr.Name, ts.Name),
		Namespace: cr.Namespace,
	}

	job := k8s.NewPerfJob(objectMeta, "kafkaclient", cr.Spec.Image, cr.Spec.PodConfig)
	job.Spec.Parallelism = &ts.Threads

	initContainer := corev1.Container{
		Name:            "kafka-producer-init",
		Image:           cr.Spec.Image.Name,
		ImagePullPolicy: corev1.PullPolicy(cr.Spec.Image.PullPolicy),
		Command:         []string{"/bin/sh"},
		Args:            ProducerInitJobArgs(cr, ts),
		Resources:       cr.Spec.PodConfig.Resources,
	}
	job.Spec.Template.Spec.InitContainers = append(job.Spec.Template.Spec.InitContainers, initContainer)

	job.Spec.Template.Spec.Containers[0].Command = []string{"/bin/sh"}
	job.Spec.Template.Spec.Containers[0].Args = ProducerJobCmd(cr, ts)

	return job
}

func ProducerInitJobArgs(cr *perfv1alpha1.KafkaBench, ts *perfv1alpha1.KafkaTestSpec) []string {
	return []string{
		"/usr/bin/kafka-topics",
		"--zookeeper", strings.Join(cr.Spec.ZooKeepers, ","),
		"--create",
		"--topic", fmt.Sprintf("%s-%s-bench", cr.Name, ts.Name),
		"--partitions", fmt.Sprintf("%d", ts.Partitions),
		"--replication-factor", fmt.Sprintf("%d", ts.Replication),
		"--if-not-exists",
	}
}

func ProducerJobCmd(cr *perfv1alpha1.KafkaBench, ts *perfv1alpha1.KafkaTestSpec) []string {
	return append([]string{
		"/usr/bin/kafka-producer-perf-test",
		"--topic", fmt.Sprintf("%s-%s-bench", cr.Name, ts.Name),
		"--num-records", fmt.Sprintf("%d", ts.Records),
		"--throughput", "-1",
		"--record-size", fmt.Sprintf("%d", ts.RecordSize),
		"--producer-props", fmt.Sprintf("bootstrap.servers=%s", strings.Join(cr.Spec.Brokers, ",")),
	}, ts.ExtraProducerOpts...)
}
