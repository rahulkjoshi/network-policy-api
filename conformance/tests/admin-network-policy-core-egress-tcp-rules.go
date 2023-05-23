/*
Copyright 2022 The Kubernetes Authors.

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

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/test/e2e/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/network-policy-api/apis/v1alpha1"
	"sigs.k8s.io/network-policy-api/conformance/utils/kubernetes"
	"sigs.k8s.io/network-policy-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests,
		AdminNetworkPolicyEgressTCP,
	)
}

var AdminNetworkPolicyEgressTCP = suite.ConformanceTest{
	ShortName:   "AdminNetworkPolicyEgressTCP",
	Description: "Tests support for egress traffic (TCP protocol) using admin network policy API based on a server and client model",
	Features: []suite.SupportedFeature{
		suite.SupportAdminNetworkPolicy,
	},
	Manifests: []string{"tests/admin-network-policy-core-egress-tcp-rules_base.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {

		t.Run("Should support an 'allow-egress' policy for TCP protocol; ensure rule ordering is respected", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `egress-tcp` ANP
			// luna-lovegood-0 is our server pod in ravenclaw namespace
			clientPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-ravenclaw",
				Name:      "luna-lovegood-0",
			}, clientPod)
			framework.ExpectNoError(err, "unable to fetch the server pod")
			// harry-potter-0 is our client pod in gryffindor namespace
			// ensure egress is ALLOWED to ravenclaw from gryffindor
			// egressRule at index0 will take precedence over egressRule at index1; thus ALLOW takes precedence over DENY since rules are ordered
			success := kubernetes.PokeServer(t, "network-policy-conformance-gryffindor", "harry-potter-0", "tcp",
				clientPod.Status.PodIP, int32(80), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
			success = kubernetes.PokeServer(t, "network-policy-conformance-gryffindor", "harry-potter-1", "tcp",
				clientPod.Status.PodIP, int32(8080), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
		})

		t.Run("Should support an 'allow-egress' policy for TCP protocol at the specified port", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `egress-tcp` ANP
			// cedric-diggory-1 is our server pod in hufflepuff namespace
			clientPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-hufflepuff",
				Name:      "cedric-diggory-1",
			}, clientPod)
			framework.ExpectNoError(err, "unable to fetch the server pod")
			// harry-potter-0 is our client pod in gryffindor namespace
			// ensure egress is ALLOWED to hufflepuff from gryffindor at port 80; egressRule at index5
			success := kubernetes.PokeServer(t, "network-policy-conformance-gryffindor", "harry-potter-0", "tcp",
				clientPod.Status.PodIP, int32(8080), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
			// harry-potter-1 is our client pod in gryffindor namespace
			// ensure egress is DENIED to hufflepuff from gryffindor for rest of the traffic; egressRule at index6
			success = kubernetes.PokeServer(t, "network-policy-conformance-gryffindor", "harry-potter-1", "tcp",
				clientPod.Status.PodIP, int32(80), s.TimeoutConfig.RequestTimeout, false)
			assert.Equal(t, true, success)
		})

		t.Run("Should support an 'deny-egress' policy for TCP protocol; ensure rule ordering is respected", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `egress-tcp` ANP
			// luna-lovegood-1 is our server pod in ravenclaw namespace
			clientPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-ravenclaw",
				Name:      "luna-lovegood-1",
			}, clientPod)
			framework.ExpectNoError(err, "unable to fetch the server pod")
			anp := &v1alpha1.AdminNetworkPolicy{}
			err = s.Client.Get(ctx, client.ObjectKey{
				Name: "egress-tcp",
			}, anp)
			framework.ExpectNoError(err, "unable to fetch the admin network policy")
			// swap rules at index0 and index1
			allowRule := anp.DeepCopy().Spec.Egress[0]
			anp.Spec.Egress[0] = anp.DeepCopy().Spec.Egress[1]
			anp.Spec.Egress[1] = allowRule
			err = s.Client.Update(ctx, anp)
			framework.ExpectNoError(err, "unable to update the admin network policy")
			// harry-potter-0 is our client pod in gryffindor namespace
			// ensure egress is DENIED to ravenclaw from gryffindor
			// egressRule at index0 will take precedence over egressRule at index1; thus DENY takes precedence over ALLOW since rules are ordered
			success := kubernetes.PokeServer(t, "network-policy-conformance-gryffindor", "harry-potter-0", "tcp",
				clientPod.Status.PodIP, int32(80), s.TimeoutConfig.RequestTimeout, false)
			assert.Equal(t, true, success)
			// harry-potter-1 is our client pod in gryffindor namespace
			success = kubernetes.PokeServer(t, "network-policy-conformance-gryffindor", "harry-potter-1", "tcp",
				clientPod.Status.PodIP, int32(8080), s.TimeoutConfig.RequestTimeout, false)
			assert.Equal(t, true, success)
		})

		t.Run("Should support a 'deny-egress' policy for TCP protocol at the specified port", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `egress-tcp` ANP
			// draco-malfoy-0 is our server pod in slytherin namespace
			clientPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-slytherin",
				Name:      "draco-malfoy-0",
			}, clientPod)
			framework.ExpectNoError(err, "unable to fetch the server pod")
			// harry-potter-0 is our client pod in gryffindor namespace
			// ensure egress to slytherin is DENIED from gryffindor at port 80; egressRule at index3
			success := kubernetes.PokeServer(t, "network-policy-conformance-gryffindor", "harry-potter-0", "tcp",
				clientPod.Status.PodIP, int32(80), s.TimeoutConfig.RequestTimeout, false)
			assert.Equal(t, true, success)
			// harry-potter-1 is our client pod in gryffindor namespace
			// ensure egress to slytherin is ALLOWED from gryffindor for rest of the traffic; matches no rules hence allowed
			success = kubernetes.PokeServer(t, "network-policy-conformance-gryffindor", "harry-potter-1", "tcp",
				clientPod.Status.PodIP, int32(8080), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
		})

		t.Run("Should support an 'pass-egress' policy for TCP protocol; ensure rule ordering is respected", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `egress-tcp` ANP
			// luna-lovegood-0 is our server pod in ravenclaw namespace
			clientPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-ravenclaw",
				Name:      "luna-lovegood-0",
			}, clientPod)
			framework.ExpectNoError(err, "unable to fetch the server pod")
			anp := &v1alpha1.AdminNetworkPolicy{}
			err = s.Client.Get(ctx, client.ObjectKey{
				Name: "egress-tcp",
			}, anp)
			framework.ExpectNoError(err, "unable to fetch the admin network policy")
			// swap rules at index0 and index2
			denyRule := anp.DeepCopy().Spec.Egress[0]
			anp.Spec.Egress[0] = anp.DeepCopy().Spec.Egress[2]
			anp.Spec.Egress[2] = denyRule
			err = s.Client.Update(ctx, anp)
			framework.ExpectNoError(err, "unable to update the admin network policy")
			// harry-potter-0 is our server pod in gryffindor namespace
			// ensure egress is PASSED from gryffindor to ravenclaw
			// egressRule at index0 will take precedence over egressRule at index1&index2; thus PASS takes precedence over ALLOW/DENY since rules are ordered
			success := kubernetes.PokeServer(t, "network-policy-conformance-gryffindor", "harry-potter-0", "tcp",
				clientPod.Status.PodIP, int32(80), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
			// harry-potter-1 is our server pod in gryffindor namespace
			success = kubernetes.PokeServer(t, "network-policy-conformance-gryffindor", "harry-potter-1", "tcp",
				clientPod.Status.PodIP, int32(8080), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
		})

		t.Run("Should support a 'pass-egress' policy for TCP protocol at the specified port", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `egress-tcp` ANP
			// draco-malfoy-0 is our server pod in slytherin namespace
			clientPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-slytherin",
				Name:      "draco-malfoy-0",
			}, clientPod)
			framework.ExpectNoError(err, "unable to fetch the server pod")
			anp := &v1alpha1.AdminNetworkPolicy{}
			err = s.Client.Get(ctx, client.ObjectKey{
				Name: "egress-tcp",
			}, anp)
			framework.ExpectNoError(err, "unable to fetch the admin network policy")
			// swap rules at index3 and index4
			denyRule := anp.DeepCopy().Spec.Egress[3]
			anp.Spec.Egress[3] = anp.DeepCopy().Spec.Egress[4]
			anp.Spec.Egress[4] = denyRule
			err = s.Client.Update(ctx, anp)
			framework.ExpectNoError(err, "unable to update the admin network policy")
			// harry-potter-0 is our client pod in gryffindor namespace
			// ensure egress from gryffindor is PASSED to slytherin at port 80; egressRule at index3
			success := kubernetes.PokeServer(t, "network-policy-conformance-gryffindor", "harry-potter-0", "tcp",
				clientPod.Status.PodIP, int32(80), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
			// harry-potter-1 is our client pod in gryffindor namespace
			// ensure egress from gryffindor is ALLOWED to slytherin for rest of the traffic; matches no rules hence allowed
			success = kubernetes.PokeServer(t, "network-policy-conformance-gryffindor", "harry-potter-1", "tcp",
				clientPod.Status.PodIP, int32(8080), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
		})
	},
}