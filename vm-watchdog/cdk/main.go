package main

import (
	"cdk.tf/go/stack/generated/yandex-cloud/yandex/functioniambinding"
	"cdk.tf/go/stack/generated/yandex-cloud/yandex/functionresource"
	"cdk.tf/go/stack/generated/yandex-cloud/yandex/functiontrigger"
	"cdk.tf/go/stack/generated/yandex-cloud/yandex/iamserviceaccount"
	"cdk.tf/go/stack/generated/yandex-cloud/yandex/provider"
	rfm "cdk.tf/go/stack/generated/yandex-cloud/yandex/resourcemanagerfolderiammember"
	"fmt"
	"github.com/aws/constructs-go/constructs/v10"
	j "github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"os"
	"path"
)

func NewVmWatchdogStack(scope constructs.Construct, id string) cdktf.TerraformStack {
	stack := cdktf.NewTerraformStack(scope, &id)

	vmId := cdktf.NewTerraformVariable(stack, j.String("vmId"), &cdktf.TerraformVariableConfig{
		Default:  nil,
		Nullable: j.Bool(false),
		Type:     cdktf.VariableType_STRING(),
	})
	turnOnSchedule := cdktf.NewTerraformVariable(stack, j.String("vmId"), &cdktf.TerraformVariableConfig{
		Default:  "8 0 ? * * *",
		Nullable: j.Bool(false),
		Type:     cdktf.VariableType_STRING(),
	})
	turnOffSchedule := cdktf.NewTerraformVariable(stack, j.String("vmId"), &cdktf.TerraformVariableConfig{
		Default:  "18 0 ? * * *",
		Nullable: j.Bool(false),
		Type:     cdktf.VariableType_STRING(),
	})

	provider.NewYandexProvider(stack, j.String("provider"), &provider.YandexProviderConfig{})

	triggerManager := iamserviceaccount.NewIamServiceAccount(stack, j.String("trigger-manager"),
		&iamserviceaccount.IamServiceAccountConfig{
			Name: j.String("trigger-manager"),
		})
	functionManager := iamserviceaccount.NewIamServiceAccount(stack, j.String("function-manager"),
		&iamserviceaccount.IamServiceAccountConfig{
			Name: j.String("function-manager"),
		})

	rfm.NewResourcemanagerFolderIamMember(stack, j.String("function-vm-operator"),
		&rfm.ResourcemanagerFolderIamMemberConfig{
			Role:   j.String("compute.operator"),
			Member: j.String(fmt.Sprintf("serviceAccount:%s", *functionManager.Id())),
		})

	rfm.NewResourcemanagerFolderIamMember(stack, j.String("function-vm-viewer"),
		&rfm.ResourcemanagerFolderIamMemberConfig{
			Role:   j.String("viewer"),
			Member: j.String(fmt.Sprintf("serviceAccount:%s", *functionManager.Id())),
		})

	cwd, _ := os.Getwd()

	asset := cdktf.NewTerraformAsset(stack, j.String("vm-watchdog-asset"), &cdktf.TerraformAssetConfig{
		Path: j.String(path.Join(cwd, "../src")),
		Type: cdktf.AssetType_ARCHIVE,
	})

	vmWatchdogFunc := functionresource.NewFunctionResource(stack, j.String("vm-watchdog"), &functionresource.FunctionResourceConfig{
		Content: &functionresource.FunctionResourceContent{
			ZipFilename: asset.Path(),
		},
		Entrypoint:       j.String("vm-watchdog.InstanceHandler"),
		ExecutionTimeout: j.String("300"),
		Memory:           j.Number(128),
		Name:             j.String("wathcdog"),
		Runtime:          j.String("golang116"),
		ServiceAccountId: functionManager.Id(),
		UserHash:         asset.AssetHash(),
		Tags:             j.Strings("start", "stop"),
		Environment: &map[string]*string{
			"INSTANCE_ID": vmId.StringValue(),
		},
	})

	functioniambinding.NewFunctionIamBinding(stack, j.String("trigger-invoker"), &functioniambinding.FunctionIamBindingConfig{
		FunctionId: vmWatchdogFunc.Id(),
		Role:       j.String("serverless.functions.invoker"),
		Members: j.Strings(
			fmt.Sprintf("serviceAccount:%s", *triggerManager.Id()),
		),
	})

	functiontrigger.NewFunctionTrigger(stack, j.String("start-trigger"), &functiontrigger.FunctionTriggerConfig{
		Function: &functiontrigger.FunctionTriggerFunction{
			Id:               vmWatchdogFunc.Id(),
			ServiceAccountId: triggerManager.Id(),
			Tag:              j.String("start"),
		},
		Name: j.String("start-trigger"),
		Timer: &functiontrigger.FunctionTriggerTimer{
			CronExpression: turnOnSchedule.StringValue(),
		},
	})

	functiontrigger.NewFunctionTrigger(stack, j.String("stop-trigger"), &functiontrigger.FunctionTriggerConfig{
		Function: &functiontrigger.FunctionTriggerFunction{
			Id:               vmWatchdogFunc.Id(),
			ServiceAccountId: triggerManager.Id(),
			Tag:              j.String("stop"),
		},
		Name: j.String("stop-trigger"),
		Timer: &functiontrigger.FunctionTriggerTimer{
			CronExpression: turnOffSchedule.StringValue(),
		},
	})

	return stack
}

func main() {
	app := cdktf.NewApp(nil)

	NewVmWatchdogStack(app, "cdk")

	app.Synth()
}
