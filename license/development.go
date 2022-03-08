package license

import (
	"github.com/destructiqn/kogtevran/generic"
)

type DevelopmentLicense struct{}

func (d *DevelopmentLicense) IsRelated(_ generic.Tunnel) bool {
	return true
}

func (d *DevelopmentLicense) GetFeatures() uint64 {
	return 0xFFFFFFFFFFFFFFFF
}

func (d *DevelopmentLicense) HasFeature(_ Feature) bool {
	return true
}
