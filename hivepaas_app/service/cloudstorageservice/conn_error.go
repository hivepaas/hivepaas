package cloudstorageservice

import (
	"errors"
	"net"
	"net/http"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// ConnError words what an S3 request's failure says about a storage's settings,
// for a person testing them: the endpoint, the key or the bucket. The S3 error
// stays the cause, for the log.
func ConnError(err error) error {
	if err == nil {
		return nil
	}
	return hperrors.Wrap(hperrors.ErrCloudStorageConnFailed).WithParam("Reason", connReason(err)).WithCause(err)
}

func connReason(err error) string {
	// The SDK's response errors, reached without importing the SDK's transport:
	// what they have in common is the status the store answered with.
	var statusErr interface{ HTTPStatusCode() int }
	if errors.As(err, &statusErr) {
		// A HEAD has no body, so the status is all there is, and stores use it
		// differently: R2 answers a wrong key 400 and a bucket the key may not
		// see 403. The reasons name everything a status can mean.
		switch code := statusErr.HTTPStatusCode(); code {
		case http.StatusUnauthorized, http.StatusForbidden:
			return "access was refused - check the key auth's key ID and secret key, the bucket's name, " +
				"and that the key may read that bucket"
		case http.StatusNotFound:
			return "no such bucket at this endpoint - check the bucket's name"
		case http.StatusMovedPermanently, http.StatusBadRequest:
			return "the request was rejected - check the key auth's key ID and secret key, the region " +
				"and the endpoint"
		case 0:
		default:
			return "the store answered " + http.StatusText(code)
		}
	}
	var dnsErr *net.DNSError
	var opErr *net.OpError
	if errors.As(err, &dnsErr) || errors.As(err, &opErr) {
		return "the endpoint could not be reached - check its address"
	}
	return "the request failed - check the endpoint, the region and the key"
}
