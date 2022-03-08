package license

import "github.com/destructiqn/kogtevran/generic"

type License interface {
	GetFeatures() uint64
	HasFeature(feature Feature) bool
	IsRelated(tunnel generic.Tunnel) bool
}
