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

package application

import (
	"context"
	"strconv"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"

	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	zpa "github.com/haarchri/zpa-go-client/pkg/client"
	"github.com/haarchri/zpa-go-client/pkg/client/application_controller"
	"github.com/haarchri/zpa-go-client/pkg/models"

	v1alpha1 "github.com/haarchri/provider-zpa/apis/application/v1alpha1"
	zpaclient "github.com/haarchri/provider-zpa/pkg/client"
)

const (
	errNotApplication = "managed resource is not an Application custom resource"
	errCreateFailed   = "cannot create Application"
	errDescribeFailed = "cannot describe Application"

// errUpdateFailed                = "cannot update Application custom resource"
// errDeleteFailed                = "cannot delete Application"
// errIsUpToDateFailed            = "isUpToDate failed"
// errGetSelectors                = "cannot get system selectors"
// errGetSelectorsInvalidResponse = "get system selectors returned an unexpected response"
// errCompareSelectors            = "cannot compare selectors"
// errUpdateSelectors             = "cannotUpdateselectors"
)

// SetupApplication adds a controller that reconciles Applications.
func SetupApplication(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.ApplicationGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1alpha1.Application{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.ApplicationGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: zpa.New}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(transport runtime.ClientTransport, formats strfmt.Registry) *zpa.ZscalerPrivateAccessAPIPortal
}

type external struct {
	client *zpa.ZscalerPrivateAccessAPIPortal
	kube   client.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*v1alpha1.Application)
	if !ok {
		return nil, errors.New(errNotApplication)
	}

	cfg, err := zpaclient.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}

	client := c.newClientFn(cfg, strfmt.Default)
	return &external{client, c.kube}, nil
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Application)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotApplication)
	}

	id := meta.GetExternalName(cr)
	if id == "" {
		return managed.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: false,
		}, nil
	}

	// string to int64
	applicationID, _ := strconv.ParseInt(id, 10, 64)

	req := &application_controller.GetApplicationUsingGET1Params{
		Context:       ctx,
		ApplicationID: applicationID,
		CustomerID:    cr.Spec.ForProvider.CustomerID,
	}
	resp, reqErr := e.client.ApplicationController.GetApplicationUsingGET1(req)
	if reqErr != nil {
		return managed.ExternalObservation{ResourceExists: false}, errors.Wrap(resource.Ignore(IsNotFound, reqErr), errDescribeFailed)
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	e.LateInitialize(cr, resp)

	cr.Status.SetConditions(v1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        true,
		ResourceLateInitialized: !cmp.Equal(&cr.Spec.ForProvider, currentSpec),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Application)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotApplication)
	}

	req := &application_controller.AddApplicationUsingPOST1Params{
		Context:    ctx,
		CustomerID: cr.Spec.ForProvider.CustomerID,
		Application: &models.ApplicationResource{
			BypassType:           cr.Spec.ForProvider.BypassType,
			ConfigSpace:          cr.Spec.ForProvider.ConfigSpace,
			DefaultIdleTimeout:   cr.Spec.ForProvider.DefaultIdleTimeout,
			DefaultMaxAge:        cr.Spec.ForProvider.DefaultMaxAge,
			Description:          cr.Spec.ForProvider.Description,
			DomainNames:          cr.Spec.ForProvider.DomainNames,
			DoubleEncrypt:        cr.Spec.ForProvider.DoubleEncrypt,
			Enabled:              cr.Spec.ForProvider.Enabled,
			HealthCheckType:      cr.Spec.ForProvider.HealthCheckType,
			HealthReporting:      cr.Spec.ForProvider.HealthReporting,
			IcmpAccessType:       cr.Spec.ForProvider.IcmpAccessType,
			IPAnchored:           cr.Spec.ForProvider.IPAnchored,
			IsCnameEnabled:       cr.Spec.ForProvider.IsCnameEnabled,
			Name:                 cr.Name,
			PassiveHealthEnabled: cr.Spec.ForProvider.PassiveHealthEnabled,
			SegmentGroupID:       cr.Spec.ForProvider.SegmentGroupID,
			SegmentGroupName:     cr.Spec.ForProvider.SegmentGroupName,
			TCPPortRanges:        cr.Spec.ForProvider.TCPPortRanges,
			UDPPortRanges:        cr.Spec.ForProvider.UDPPortRanges,
		},
	}

	resp, err := e.client.ApplicationController.AddApplicationUsingPOST1(req)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	meta.SetExternalName(cr, *zpaclient.String(resp.Payload.ID))
	return managed.ExternalCreation{
		ExternalNameAssigned: true,
	}, nil

}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	return nil
}

func (e *external) LateInitialize(cr *v1alpha1.Application, obj *application_controller.GetApplicationUsingGET1OK) { // nolint:gocyclo

	// if cr.Spec.ForProvider.CreationTime != obj.Payload.CreationTime {
	// 	cr.Spec.ForProvider.CreationTime = obj.Payload.CreationTime
	// }

	if cr.Spec.ForProvider.ModifiedBy != obj.Payload.ModifiedBy {
		cr.Spec.ForProvider.ModifiedBy = obj.Payload.ModifiedBy
	}

	if cr.Spec.ForProvider.Enabled != obj.Payload.Enabled {
		cr.Spec.ForProvider.Enabled = obj.Payload.Enabled
	}

	if cr.Spec.ForProvider.PassiveHealthEnabled != obj.Payload.PassiveHealthEnabled {
		cr.Spec.ForProvider.PassiveHealthEnabled = obj.Payload.PassiveHealthEnabled
	}

	if cr.Spec.ForProvider.DoubleEncrypt != obj.Payload.DoubleEncrypt {
		cr.Spec.ForProvider.DoubleEncrypt = obj.Payload.DoubleEncrypt
	}

	if cr.Spec.ForProvider.ConfigSpace != obj.Payload.ConfigSpace {
		cr.Spec.ForProvider.ConfigSpace = obj.Payload.ConfigSpace
	}

	if cr.Spec.ForProvider.BypassType != obj.Payload.BypassType {
		cr.Spec.ForProvider.BypassType = obj.Payload.BypassType
	}

	if cr.Spec.ForProvider.HealthCheckType != obj.Payload.HealthCheckType {
		cr.Spec.ForProvider.HealthCheckType = obj.Payload.HealthCheckType
	}

	if cr.Spec.ForProvider.IcmpAccessType != obj.Payload.IcmpAccessType {
		cr.Spec.ForProvider.IcmpAccessType = obj.Payload.IcmpAccessType
	}

	if cr.Spec.ForProvider.IsCnameEnabled != obj.Payload.IsCnameEnabled {
		cr.Spec.ForProvider.IsCnameEnabled = obj.Payload.IsCnameEnabled
	}

	if cr.Spec.ForProvider.IPAnchored != obj.Payload.IPAnchored {
		cr.Spec.ForProvider.IPAnchored = obj.Payload.IPAnchored
	}

	if cr.Spec.ForProvider.SegmentGroupName != obj.Payload.SegmentGroupName {
		cr.Spec.ForProvider.SegmentGroupName = obj.Payload.SegmentGroupName
	}

}
