package storage

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	storagev1beta1 "github.com/gamefabric/gf-core/pkg/api/storage/v1beta1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestNewVolumeStoreRetentionPolicyModel(t *testing.T) {
	model := newVolumeStoreRetentionPolicyModel(testVolumeRetentionPolicyObject)

	assert.Equal(t, testVolumeStoreRetentionPolicyModel, model)
}

func TestVolumeStoreRetentionPolicyModel_ToObject(t *testing.T) {
	obj := testVolumeStoreRetentionPolicyModel.ToObject()

	assert.Equal(t, testVolumeRetentionPolicyObject, obj)
}

var (
	testVolumeRetentionPolicyObject = &storagev1beta1.VolumeStoreRetentionPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-volume",
			Environment: "dflt",
		},
		Spec: storagev1beta1.VolumeStoreRetentionPolicySpec{
			VolumeStoreName: "test-volume-store",
			Snapshots: &storagev1beta1.VolumeStoreRetentionPolicySnapshots{
				OfflineCount: 1,
				OfflineDays:  2,
				OnlineCount:  3,
				OnlineDays:   4,
			},
		},
	}

	testVolumeStoreRetentionPolicyModel = volumeStoreRetentionPolicyModel{
		ID:          types.StringValue("dflt/test-volume"),
		Name:        types.StringValue("test-volume"),
		Environment: types.StringValue("dflt"),
		VolumeStore: types.StringValue("test-volume-store"),
		OfflineSnapshot: &volumeStoreRetentionPolicySnapshot{
			Count: types.Int64Value(1),
			Days:  types.Int64Value(2),
		},
		OnlineSnapshot: &volumeStoreRetentionPolicySnapshot{
			Count: types.Int64Value(3),
			Days:  types.Int64Value(4),
		},
	}
)
