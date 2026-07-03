package wait

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/gamefabric/gf-apiclient/tools/clientset"
	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	v1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-apicore/runtime"
)

// PollUntilNotFound polls until the resource is not found.
func PollUntilNotFound[T runtime.Object](ctx context.Context, getFn clientset.Getter[T], name string) error {
	bo := backoff.NewExponentialBackOff()
	opts := []backoff.RetryOption{
		backoff.WithBackOff(bo),
		backoff.WithMaxElapsedTime(24 * time.Hour), // README: Poll forever until the object is gone. We can revisit this decision if necessary.
	}

	_, err := backoff.Retry(ctx, func() (struct{}, error) {
		_, err := getFn.Get(ctx, name, v1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return struct{}{}, nil
			}
			return struct{}{}, err
		}
		return struct{}{}, fmt.Errorf("%q still present", name)
	}, opts...)

	return err
}
