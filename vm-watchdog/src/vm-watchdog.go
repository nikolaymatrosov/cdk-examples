package main

import (
	"context"
	"fmt"
	"os"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/triggers/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

func InstanceHandler(ctx context.Context, event TimerTriggerEvent) (*Response, error) { //nolint: deadcode,unused
	instanceID := os.Getenv("INSTANCE_ID")

	sdk, err := ycsdk.Build(ctx, ycsdk.Config{
		Credentials: ycsdk.InstanceServiceAccount(),
	})
	if err != nil {
		return nil, err
	}

	triggerID := event.Messages[0].Details.TriggerID

	trigger, err := sdk.Serverless().Triggers().Trigger().Get(ctx, &triggers.GetTriggerRequest{TriggerId: triggerID})
	if err != nil {
		return nil, err
	}
	function := trigger.Rule.GetTimer().GetInvokeFunction()

	vm, err := sdk.Compute().Instance().Get(ctx, &compute.GetInstanceRequest{InstanceId: instanceID})
	if err != nil {
		return nil, err
	}

	if function.FunctionTag == "start" {
		if vm.Status == compute.Instance_STOPPED || vm.Status == compute.Instance_STOPPING {
			op, err := sdk.WrapOperation(sdk.Compute().Instance().Start(ctx, &compute.StartInstanceRequest{
				InstanceId: instanceID,
			}))
			if err != nil {
				return nil, err
			}

			if opErr := op.Error(); opErr != nil {
				return &Response{
					StatusCode: 200,
					Body: fmt.Sprintf(
						"Failed to start VM: %s",
						op.Error()),
				}, nil
			}

			meta, err := op.Metadata()
			if err != nil {
				return nil, err
			}

			return &Response{
				StatusCode: 200,
				Body: fmt.Sprintf("Instance %s started",
					meta.(*compute.StartInstanceMetadata).GetInstanceId(),
				),
			}, nil
		}
		return &Response{
			StatusCode: 200,
			Body:       "Failed to start instance: already started or in invalid state",
		}, nil
	}
	if function.FunctionTag == "stop" {
		if vm.Status == compute.Instance_RUNNING {
			op, err := sdk.WrapOperation(sdk.Compute().Instance().Stop(ctx, &compute.StopInstanceRequest{
				InstanceId: instanceID,
			}))
			if err != nil {
				return nil, err
			}

			if opErr := op.Error(); opErr != nil {
				return &Response{
					StatusCode: 200,
					Body: fmt.Sprintf(
						"Failed to stop VM: %s",
						op.Error()),
				}, nil
			}

			meta, err := op.Metadata()
			if err != nil {
				return nil, err
			}

			return &Response{
				StatusCode: 200,
				Body: fmt.Sprintf("Instance %s stopped",
					meta.(*compute.StopInstanceMetadata).GetInstanceId(),
				),
			}, nil
		}
		return &Response{
			StatusCode: 200,
			Body:       "Failed to stop instance: already stopped or in invalid state",
		}, nil
	}
	return &Response{
		StatusCode: 200,
		Body:       fmt.Sprintf("Called by unknown trigger %v", triggerID),
	}, nil
}
