package data

import "github.com/fuserobotics/kvgossip/util"

func DedupeSignedData(da []*SignedData) (res []*SignedData) {
	seen := make(map[string]*SignedData)
	res = []*SignedData{}
	for _, sd := range da {
		key := util.HexSha256(sd.Body)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = sd
		res = append(res, sd)
	}
	return
}
