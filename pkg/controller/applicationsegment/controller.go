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

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	zpa "github.com/haarchri/zpa-go-client/pkg/client"
	"github.com/haarchri/zpa-go-client/pkg/client/application_controller"
	"github.com/haarchri/zpa-go-client/pkg/client/server_group_controller"
	"github.com/haarchri/zpa-go-client/pkg/models"

	v1alpha1 "github.com/crossplane-contrib/provider-zpa/apis/applicationsegment/v1alpha1"
	store "github.com/crossplane-contrib/provider-zpa/apis/v1alpha1"
	zpaclient "github.com/crossplane-contrib/provider-zpa/pkg/client"
	"github.com/crossplane-contrib/provider-zpa/pkg/features"
)

const (
	errNotApplicationSegment = "managed resource is not an ApplicationSegment custom resource"
	errCreateFailed          = "cannot create ApplicationSegment"
	errDescribeFailed        = "cannot describe ApplicationSegment"
	errUpdateFailed          = "cannot update ApplicationSegment"
	errDeleteFailed          = "cannot delete ApplicationSegment"
	errServerGroupNotFound   = "cannot get ServerGroup"
)

// SetupApplicationSegment adds a controller that reconciles ApplicationSegments.
func SetupApplicationSegment(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.ApplicationSegmentGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), store.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.ApplicationSegment{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.ApplicationSegmentGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: zpa.New}),
			managed.WithConnectionPublishers(),
			managed.WithInitializers(),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
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
	_, ok := mg.(*v1alpha1.ApplicationSegment)
	if !ok {
		return nil, errors.New(errNotApplicationSegment)
	}

	cfg, err := zpaclient.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}

	client := c.newClientFn(cfg, strfmt.Default)
	return &external{client, c.kube}, nil
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.ApplicationSegment)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotApplicationSegment)
	}

	id := meta.GetExternalName(cr)
	if id == "" {
		return managed.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: false,
		}, nil
	}

	customerID, _ := zpaclient.CustomerID(ctx, e.kube, mg)
	req := &application_controller.GetApplicationUsingGET1Params{
		Context:       ctx,
		ApplicationID: id,
		CustomerID:    customerID,
	}
	resp, reqErr := e.client.ApplicationController.GetApplicationUsingGET1(req)
	if reqErr != nil {
		return managed.ExternalObservation{ResourceExists: false}, errors.Wrap(resource.Ignore(IsNotFound, reqErr), errDescribeFailed)
	}

	cr.Status.AtProvider = generateObservation(resp)

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	e.LateInitialize(cr, resp)

	cr.Status.SetConditions(v1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isUpToDate(&cr.Spec.ForProvider, resp),
		ResourceLateInitialized: !cmp.Equal(&cr.Spec.ForProvider, currentSpec),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ApplicationSegment)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotApplicationSegment)
	}

	customerID, _ := zpaclient.CustomerID(ctx, e.kube, mg)
	req := &application_controller.AddApplicationUsingPOST1Params{
		Context:    ctx,
		CustomerID: customerID,
		Application: &models.ApplicationResource{
			BypassType:           cr.Spec.ForProvider.BypassType,
			ConfigSpace:          cr.Spec.ForProvider.ConfigSpace,
			DefaultIdleTimeout:   cr.Spec.ForProvider.DefaultIdleTimeout,
			DefaultMaxAge:        cr.Spec.ForProvider.DefaultMaxAge,
			Description:          cr.Spec.ForProvider.Description,
			DomainNames:          cr.Spec.ForProvider.DomainNames,
			DoubleEncrypt:        zpaclient.BoolValue(cr.Spec.ForProvider.DoubleEncrypt),
			Enabled:              zpaclient.BoolValue(cr.Spec.ForProvider.Enabled),
			HealthCheckType:      cr.Spec.ForProvider.HealthCheckType,
			HealthReporting:      cr.Spec.ForProvider.HealthReporting,
			IcmpAccessType:       cr.Spec.ForProvider.IcmpAccessType,
			IPAnchored:           zpaclient.BoolValue(cr.Spec.ForProvider.IPAnchored),
			IsCnameEnabled:       zpaclient.BoolValue(cr.Spec.ForProvider.IsCnameEnabled),
			Name:                 cr.Spec.ForProvider.Name,
			PassiveHealthEnabled: zpaclient.BoolValue(cr.Spec.ForProvider.PassiveHealthEnabled),
			SegmentGroupID:       zpaclient.StringValue(cr.Spec.ForProvider.SegmentGroupID),
			TCPPortRanges:        cr.Spec.ForProvider.TCPPortRanges,
			UDPPortRanges:        cr.Spec.ForProvider.UDPPortRanges,
			ServerGroups:         make([]*models.AppServerGroup, 0),
		},
	}

	for i := range cr.Spec.ForProvider.ServerGroups {
		servergroupreq := &server_group_controller.GetServerGroupUsingGET1Params{
			Context:    ctx,
			CustomerID: customerID,
			GroupID:    cr.Spec.ForProvider.ServerGroups[i],
		}
		servergroupresp, err := e.client.ServerGroupController.GetServerGroupUsingGET1(servergroupreq)
		if err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errServerGroupNotFound)
		}

		req.Application.ServerGroups = append(
			req.Application.ServerGroups,
			&models.AppServerGroup{
				ID:   cr.Spec.ForProvider.ServerGroups[i],
				Name: zpaclient.String(servergroupresp.Payload.Name),
			})
	}

	resp, err := e.client.ApplicationController.AddApplicationUsingPOST1(req)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	meta.SetExternalName(cr, resp.Payload.ID)
	return managed.ExternalCreation{
		ExternalNameAssigned: true,
	}, nil

}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.ApplicationSegment)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotApplicationSegment)
	}

	customerID, _ := zpaclient.CustomerID(ctx, e.kube, mg)
	req := &application_controller.UpdateApplicationV2UsingPUT1Params{
		Context:       ctx,
		CustomerID:    customerID,
		ApplicationID: meta.GetExternalName(cr),
		Application: &models.ApplicationResource{
			BypassType:           cr.Spec.ForProvider.BypassType,
			ConfigSpace:          cr.Spec.ForProvider.ConfigSpace,
			DefaultIdleTimeout:   cr.Spec.ForProvider.DefaultIdleTimeout,
			DefaultMaxAge:        cr.Spec.ForProvider.DefaultMaxAge,
			Description:          cr.Spec.ForProvider.Description,
			DomainNames:          cr.Spec.ForProvider.DomainNames,
			DoubleEncrypt:        zpaclient.BoolValue(cr.Spec.ForProvider.DoubleEncrypt),
			Enabled:              zpaclient.BoolValue(cr.Spec.ForProvider.Enabled),
			HealthCheckType:      cr.Spec.ForProvider.HealthCheckType,
			HealthReporting:      cr.Spec.ForProvider.HealthReporting,
			IcmpAccessType:       cr.Spec.ForProvider.IcmpAccessType,
			IPAnchored:           zpaclient.BoolValue(cr.Spec.ForProvider.IPAnchored),
			IsCnameEnabled:       zpaclient.BoolValue(cr.Spec.ForProvider.IsCnameEnabled),
			Name:                 cr.Spec.ForProvider.Name,
			PassiveHealthEnabled: zpaclient.BoolValue(cr.Spec.ForProvider.PassiveHealthEnabled),
			SegmentGroupID:       zpaclient.StringValue(cr.Spec.ForProvider.SegmentGroupID),
			TCPPortRanges:        cr.Spec.ForProvider.TCPPortRanges,
			UDPPortRanges:        cr.Spec.ForProvider.UDPPortRanges,
			ServerGroups:         make([]*models.AppServerGroup, 0),
		},
	}

	for i := range cr.Spec.ForProvider.ServerGroups {
		servergroupreq := &server_group_controller.GetServerGroupUsingGET1Params{
			Context:    ctx,
			CustomerID: customerID,
			GroupID:    cr.Spec.ForProvider.ServerGroups[i],
		}
		servergroupresp, err := e.client.ServerGroupController.GetServerGroupUsingGET1(servergroupreq)
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errServerGroupNotFound)
		}

		req.Application.ServerGroups = append(
			req.Application.ServerGroups,
			&models.AppServerGroup{
				ID:   cr.Spec.ForProvider.ServerGroups[i],
				Name: zpaclient.String(servergroupresp.Payload.Name),
			})
	}

	if _, _, err := e.client.ApplicationController.UpdateApplicationV2UsingPUT1(req); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.ApplicationSegment)
	if !ok {
		return errors.New(errNotApplicationSegment)
	}

	id := meta.GetExternalName(cr)
	if id == "" {
		return errors.New(errNotApplicationSegment)
	}

	customerID, _ := zpaclient.CustomerID(ctx, e.kube, mg)
	req := &application_controller.DeleteApplicationUsingDELETE1Params{
		Context:       ctx,
		ApplicationID: id,
		CustomerID:    customerID,
		ForceDelete:   zpaclient.Bool(true),
	}

	_, err := e.client.ApplicationController.DeleteApplicationUsingDELETE1(req)
	if err != nil {
		return errors.Wrap(err, errDeleteFailed)
	}

	return nil
}

func (e *external) LateInitialize(cr *v1alpha1.ApplicationSegment, obj *application_controller.GetApplicationUsingGET1OK) { // nolint:gocyclo

	if cr.Spec.ForProvider.Enabled == nil {
		cr.Spec.ForProvider.Enabled = zpaclient.Bool(obj.Payload.Enabled)
	}

	if cr.Spec.ForProvider.PassiveHealthEnabled == nil {
		cr.Spec.ForProvider.PassiveHealthEnabled = zpaclient.Bool(obj.Payload.PassiveHealthEnabled)
	}

	if cr.Spec.ForProvider.DoubleEncrypt == nil {
		cr.Spec.ForProvider.DoubleEncrypt = zpaclient.Bool(obj.Payload.DoubleEncrypt)
	}

	if cr.Spec.ForProvider.ConfigSpace == "" {
		cr.Spec.ForProvider.ConfigSpace = obj.Payload.ConfigSpace
	}

	if cr.Spec.ForProvider.BypassType == "" {
		cr.Spec.ForProvider.BypassType = obj.Payload.BypassType
	}

	if cr.Spec.ForProvider.HealthCheckType == "" {
		cr.Spec.ForProvider.HealthCheckType = obj.Payload.HealthCheckType
	}

	if cr.Spec.ForProvider.IcmpAccessType == "" {
		cr.Spec.ForProvider.IcmpAccessType = obj.Payload.IcmpAccessType
	}

	if cr.Spec.ForProvider.IsCnameEnabled == nil {
		cr.Spec.ForProvider.IsCnameEnabled = zpaclient.Bool(obj.Payload.IsCnameEnabled)
	}

	if cr.Spec.ForProvider.IPAnchored == nil {
		cr.Spec.ForProvider.IPAnchored = zpaclient.Bool(obj.Payload.IPAnchored)
	}

	if cr.Spec.ForProvider.HealthReporting == "" {
		cr.Spec.ForProvider.HealthReporting = obj.Payload.HealthReporting
	}

}

// generateObservation generates observation for the input object application_controller.GetApplicationUsingGET1OK
func generateObservation(in *application_controller.GetApplicationUsingGET1OK) v1alpha1.Observation {
	cr := v1alpha1.Observation{}
	obj := in.Payload

	cr.ID = obj.ID
	cr.CreationTime = obj.CreationTime
	cr.ModifiedBy = obj.ModifiedBy
	cr.ModifiedTime = obj.ModifiedTime
	cr.ApplicationSegment.BypassType = obj.BypassType
	cr.ApplicationSegment.ConfigSpace = obj.ConfigSpace
	cr.ApplicationSegment.DefaultIdleTimeout = obj.DefaultIdleTimeout
	cr.ApplicationSegment.DefaultMaxAge = obj.DefaultMaxAge
	cr.ApplicationSegment.DomainNames = append(cr.ApplicationSegment.DomainNames, obj.DomainNames...)
	cr.ApplicationSegment.DoubleEncrypt = zpaclient.Bool(obj.DoubleEncrypt)
	cr.ApplicationSegment.Enabled = zpaclient.Bool(obj.Enabled)
	cr.ApplicationSegment.HealthCheckType = obj.HealthCheckType
	cr.ApplicationSegment.HealthReporting = obj.HealthReporting
	cr.ApplicationSegment.IPAnchored = zpaclient.Bool(obj.IPAnchored)
	for i := range obj.ServerGroups {
		cr.ServerGroup = append(cr.ServerGroup, v1alpha1.AppServerGroup(*obj.ServerGroups[i]))
	}

	return cr
}

// isUpToDate checks whether there is a change in any of the modifiable fields.
func isUpToDate(cr *v1alpha1.ApplicationSegmentParameters, gobj *application_controller.GetApplicationUsingGET1OK) bool { // nolint:gocyclo
	obj := gobj.Payload

	if !zpaclient.IsEqualString(zpaclient.StringToPtr(cr.BypassType), zpaclient.StringToPtr(obj.BypassType)) {
		return false
	}

	if !zpaclient.IsEqualString(zpaclient.StringToPtr(cr.ConfigSpace), zpaclient.StringToPtr(obj.ConfigSpace)) {
		return false
	}

	if !zpaclient.IsEqualString(zpaclient.StringToPtr(cr.DefaultIdleTimeout), zpaclient.StringToPtr(obj.DefaultIdleTimeout)) {
		return false
	}

	if !zpaclient.IsEqualString(zpaclient.StringToPtr(cr.DefaultMaxAge), zpaclient.StringToPtr(obj.DefaultMaxAge)) {
		return false
	}

	if !zpaclient.IsEqualString(zpaclient.StringToPtr(cr.Description), zpaclient.StringToPtr(obj.Description)) {
		return false
	}

	if !zpaclient.IsEqualBool(cr.DoubleEncrypt, zpaclient.Bool(obj.DoubleEncrypt)) {
		return false
	}

	if !zpaclient.IsEqualBool(cr.Enabled, zpaclient.Bool(obj.Enabled)) {
		return false
	}

	if !zpaclient.IsEqualString(zpaclient.StringToPtr(cr.HealthCheckType), zpaclient.StringToPtr(obj.HealthCheckType)) {
		return false
	}

	if !zpaclient.IsEqualString(zpaclient.StringToPtr(cr.HealthReporting), zpaclient.StringToPtr(obj.HealthReporting)) {
		return false
	}

	if !zpaclient.IsEqualBool(cr.IPAnchored, zpaclient.Bool(obj.IPAnchored)) {
		return false
	}

	if !zpaclient.IsEqualString(zpaclient.StringToPtr(cr.IcmpAccessType), zpaclient.StringToPtr(obj.IcmpAccessType)) {
		return false
	}

	if !zpaclient.IsEqualBool(cr.IsCnameEnabled, zpaclient.Bool(obj.IsCnameEnabled)) {
		return false
	}

	if !zpaclient.IsEqualBool(cr.PassiveHealthEnabled, zpaclient.Bool(obj.PassiveHealthEnabled)) {
		return false
	}

	if !zpaclient.IsEqualStringArrayContent(cr.DomainNames, obj.DomainNames) {
		return false
	}

	if !zpaclient.IsEqualStringArrayContent(cr.TCPPortRanges, obj.TCPPortRanges) {
		return false
	}

	if !zpaclient.IsEqualStringArrayContent(cr.UDPPortRanges, obj.UDPPortRanges) {
		return false
	}

	if !zpaclient.IsEqualString(zpaclient.StringToPtr(cr.Name), zpaclient.StringToPtr(obj.Name)) {
		return false
	}

	return true
}
