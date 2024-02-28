package rancher

//
//import (
//	"fmt"
//	"github.com/suse-edge/edge-image-builder/pkg/image"
//	"testing"
//)
//
//func TestPre(t *testing.T) {
//	ctx, teardown := setupContext(t)
//	defer teardown()
//	ctx.Helm = helm.New(ctx.BuildDir)
//	ctx.ImageDefinition.Rancher = image.Rancher{
//		Version: "v2.8.2",
//		CertManager: image.CertManager{
//			Version: "v1.14.2",
//		},
//	}
//
//	err := ConfigureRancher(ctx)
//	if err != nil {
//		fmt.Println("error configuring rancher", err)
//	}
//}
