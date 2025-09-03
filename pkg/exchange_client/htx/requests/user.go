//nolint:tagliatelle
package requests

type GetAPIKeyInformationRequest struct {
	UID       string `json:"uid" url:"uid"`
	AccessKey string `json:"access-key,omitempty" url:"access-key,omitempty"`
}
