// Code generated : DO NOT EDIT.
// Copyright (c) 2022 Jean-Francois SMIGIELSKI
// Distributed under the MIT License

package device

import (
	"context"
	"github.com/juju/errors"
	"github.com/mydragonfly00/onvif"
	"github.com/mydragonfly00/onvif/device"
	"github.com/mydragonfly00/onvif/sdk"
)

// Call_GetCertificates forwards the call to dev.CallMethod() then parses the payload of the reply as a GetCertificatesResponse.
func Call_GetCertificates(ctx context.Context, dev *onvif.Device, request device.GetCertificates) (device.GetCertificatesResponse, error) {
	type Envelope struct {
		Header struct{}
		Body   struct {
			GetCertificatesResponse device.GetCertificatesResponse
		}
	}
	var reply Envelope
	if httpReply, err := dev.CallMethod(request); err != nil {
		return reply.Body.GetCertificatesResponse, errors.Annotate(err, "call")
	} else {
		err = sdk.ReadAndParse(ctx, httpReply, &reply, "GetCertificates")
		return reply.Body.GetCertificatesResponse, errors.Annotate(err, "reply")
	}
}
