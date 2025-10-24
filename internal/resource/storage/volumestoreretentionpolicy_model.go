package storage

import (
	"github.com/gamefabric/gf-apiclient/tools/cache"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	storagev1beta1 "github.com/gamefabric/gf-core/pkg/api/storage/v1beta1"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type volumeStoreRetentionPolicyModel struct {
	ID              types.String                        `tfsdk:"id"`
	Name            types.String                        `tfsdk:"name"`
	Environment     types.String                        `tfsdk:"environment"`
	VolumeStore     types.String                        `tfsdk:"volume_store"`
	OfflineSnapshot *volumeStoreRetentionPolicySnapshot `tfsdk:"offline_snapshot"`
	OnlineSnapshot  *volumeStoreRetentionPolicySnapshot `tfsdk:"online_snapshot"`
}

func newVolumeStoreRetentionPolicyModel(obj *storagev1beta1.VolumeStoreRetentionPolicy) volumeStoreRetentionPolicyModel {
	model := volumeStoreRetentionPolicyModel{
		ID:          types.StringValue(cache.NewObjectName(obj.Environment, obj.Name).String()),
		Name:        types.StringValue(obj.Name),
		Environment: types.StringValue(obj.Environment),
		VolumeStore: types.StringValue(obj.Spec.VolumeStoreName),
	}
	if obj.Spec.Snapshots != nil {
		model.OfflineSnapshot = &volumeStoreRetentionPolicySnapshot{
			Count: types.Int64Value(int64(obj.Spec.Snapshots.OfflineCount)),
			Days:  types.Int64Value(int64(obj.Spec.Snapshots.OfflineDays)),
		}
		model.OnlineSnapshot = &volumeStoreRetentionPolicySnapshot{
			Count: types.Int64Value(int64(obj.Spec.Snapshots.OnlineCount)),
			Days:  types.Int64Value(int64(obj.Spec.Snapshots.OnlineDays)),
		}
	}
	return model
}

func (m volumeStoreRetentionPolicyModel) ToObject() *storagev1beta1.VolumeStoreRetentionPolicy {
	obj := &storagev1beta1.VolumeStoreRetentionPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Name.ValueString(),
			Environment: m.Environment.ValueString(),
		},
		Spec: storagev1beta1.VolumeStoreRetentionPolicySpec{
			VolumeStoreName: m.VolumeStore.ValueString(),
		},
	}
	if m.OfflineSnapshot != nil || m.OnlineSnapshot != nil {
		obj.Spec.Snapshots = &storagev1beta1.VolumeStoreRetentionPolicySnapshots{}
		if m.OfflineSnapshot != nil {
			obj.Spec.Snapshots.OfflineCount = int(m.OfflineSnapshot.Count.ValueInt64())
			obj.Spec.Snapshots.OfflineDays = int(m.OfflineSnapshot.Days.ValueInt64())
		}
		if m.OnlineSnapshot != nil {
			obj.Spec.Snapshots.OnlineCount = int(m.OnlineSnapshot.Count.ValueInt64())
			obj.Spec.Snapshots.OnlineDays = int(m.OnlineSnapshot.Days.ValueInt64())
		}
	}
	return obj
}

type volumeStoreRetentionPolicySnapshot struct {
	Count types.Int64 `tfsdk:"count"`
	Days  types.Int64 `tfsdk:"days"`
}

func (s *volumeStoreRetentionPolicySnapshot) Default() defaults.Object {
	return objectdefault.StaticValue(types.ObjectValueMust(map[string]attr.Type{
		"count": types.Int64Type,
		"days":  types.Int64Type,
	}, map[string]attr.Value{
		"count": types.Int64Value(0),
		"days":  types.Int64Value(0),
	}))
}
