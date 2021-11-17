/*
Copyright 2021 The Crossplane Authors.
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

package v1alpha1

import (
	"context"

	segmentGroup "github.com/crossplane-contrib/provider-zpa/apis/segmentgroup/v1alpha1"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ResolveReferences of this ApplicationSegment
func (mg *ApplicationSegment) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.segmentGroupID
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.SegmentGroupID),
		Reference:    mg.Spec.ForProvider.SegmentGroupIDRef,
		Selector:     mg.Spec.ForProvider.SegmentGroupIDSelector,
		To:           reference.To{Managed: &segmentGroup.SegmentGroup{}, List: &segmentGroup.SegmentGroupList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.segmentGroupID")
	}
	mg.Spec.ForProvider.SegmentGroupID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.SegmentGroupIDRef = rsp.ResolvedReference

	return nil
}
