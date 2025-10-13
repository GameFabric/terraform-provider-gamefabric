package wait_test

import (
	"testing"
	"testing/synctest"
	"time"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	v1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/fake"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/wait"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPollUntilNotFound_Gone(t *testing.T) {
	t.Parallel()

	cs, err := fake.New()
	require.NoError(t, err)

	err = wait.PollUntilNotFound(t.Context(), cs.CoreV1().Environments(), "non-existent")
	require.NoError(t, err)
}

func TestPollUntilNotFound_NotGone(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {

		cs, err := fake.New(&v1.Environment{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-env",
			},
		})
		require.NoError(t, err)

		now := time.Now()

		synctest.Wait()
		err = wait.PollUntilNotFound(t.Context(), cs.CoreV1().Environments(), "test-env")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "still present")
		assert.GreaterOrEqual(t, time.Minute, time.Until(now))
	})
}

func TestPollUntilNotFound_GoneLater(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		cs, err := fake.New(&v1.Environment{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-env",
			},
		})
		require.NoError(t, err)

		now := time.Now()

		go func() {
			synctest.Wait()
			<-time.After(30 * time.Second)

			err := cs.CoreV1().Environments().Delete(t.Context(), "test-env", metav1.DeleteOptions{})
			require.NoError(t, err)
		}()

		err = wait.PollUntilNotFound(t.Context(), cs.CoreV1().Environments(), "test-env")

		require.NoError(t, err)
		assert.GreaterOrEqual(t, time.Minute, time.Until(now))
	})
}
