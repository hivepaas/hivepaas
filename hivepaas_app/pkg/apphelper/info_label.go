package apphelper

import (
	"encoding/json"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
)

type AppInfo struct {
	Name string `json:"name"`
	Key  string `json:"key"`
	Env  string `json:"env"`
}

func CalcAppInfoLabel(info *AppInfo) string {
	res, err := json.Marshal(info)
	if err != nil { // should never happen
		panic(err)
	}
	return string(res)
}

func ParseAppInfoLabel(label string) (*AppInfo, error) {
	res := &AppInfo{}
	if label == "" {
		return res, nil
	}
	err := json.Unmarshal(reflectutil.UnsafeStrToBytes(label), &res)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return res, nil
}
