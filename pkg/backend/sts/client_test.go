package sts

import (
	"testing"
	"errors"
	"fmt"
	"os"
)

func TestClient_AssumeRole(t *testing.T) {
	//url := "http://127.0.0.1:9021"
	url := os.Getenv("CRED_STS_URL")
	cli := NewStsClient(url)
	testCases := []struct {
		payload []byte
		output  error
	}{
		{
			payload: []byte(`{"role_qrn":"qrn:qingcloud:iam::usr-SdQDgc2B:role/ReatQstr","external_id":"service:ec2.qingcloud.com","instance_profile":{"instance_id":"i-test8","root_user_id":"usr-SdQDgc2B","role_qrn":"qrn:qingcloud:iam::usr-SdQDgc2B:role/ReatQstr","console_id":"admin"}}`),
			output:  nil,
		},
		{
			payload: []byte(`{"role_qrn":"qrn:qingcloud:iam::usr-SdQDgc2B:role/RunInstance","external_id":"service:ec2.qingcloud.com","instance_profile":{"instance_id":"i-test8","root_user_id":"usr-SdQDgc2B","role_qrn":"qrn:qingcloud:iam::usr-SdQDgc2B:role/RunInstance","console_id":"admin"}}`),
			output:  errors.New("AssumeRole error with status [500]"),
		},
	}
	for _, test := range testCases {
		_, err := cli.AssumeRole(test.payload)

		fmt.Printf("%v", err)

		if (err != nil && test.output == nil) || (test.output != nil && (err == nil || test.output.Error() != err.Error())) {
			t.Errorf("%v, %s", err, test.payload)
			return
		}

	}
}
