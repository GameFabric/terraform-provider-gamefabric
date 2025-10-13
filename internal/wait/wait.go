package wait

import (
	"context"
	"fmt"

	"github.com/cenkalti/backoff/v4"
	"github.com/gamefabric/gf-apiclient/tools/clientset"
	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	v1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-apicore/runtime"
)

// PollUntilNotFound polls until the resource is not found.
func PollUntilNotFound[T runtime.Object](ctx context.Context, getFn clientset.Getter[T], name string) error {
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 0 // README: Poll forever until the object is gone. We can revisit this decision if necessary.

	return backoff.Retry(func() error {
		_, err := getFn.Get(ctx, name, v1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return nil
			}
			return err
		}
		return fmt.Errorf("%q still present", name)
	}, backoff.WithContext(bo, ctx))
}
