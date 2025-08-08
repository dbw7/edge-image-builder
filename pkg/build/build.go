package build

import (
	"fmt"
	"github.com/suse-edge/edge-image-builder/pkg/context"
	"os"
	"path/filepath"

	"github.com/suse-edge/edge-image-builder/pkg/log"
)

type imageConfigurator interface {
	Configure(ctx *context.Context) error
}

type Builder struct {
	context           *context.Context
	imageConfigurator imageConfigurator
}

func NewBuilder(ctx *context.Context, imageConfigurator imageConfigurator) *Builder {
	return &Builder{
		context:           ctx,
		imageConfigurator: imageConfigurator,
	}
}

func (b *Builder) Build() error {
	log.Audit("Generating image customization components...")

	if err := b.imageConfigurator.Configure(b.context); err != nil {
		log.Audit("Error configuring customization components.")
		return fmt.Errorf("configuring image: %w", err)
	}

	switch b.context.Definition.GetImage().ImageType {
	case context.TypeISO:
		log.Audit("Building ISO image...")
		if err := b.buildIsoImage(); err != nil {
			log.Audit("Error building ISO image.")
			return err
		}
	case context.TypeRAW:
		log.Audit("Building RAW image...")
		if err := b.buildRawImage(); err != nil {
			log.Audit("Error building RAW image.")
			return err
		}
	default:
		return fmt.Errorf("invalid imageType value specified, must be either \"%s\" or \"%s\"",
			context.TypeISO, context.TypeRAW)
	}

	log.Auditf("Build complete, the image can be found at: %s",
		b.context.Definition.GetImage().OutputImageName)
	return nil
}

func (b *Builder) generateBuildDirFilename(filename string) string {
	return filepath.Join(b.context.BuildDir, filename)
}

func (b *Builder) generateOutputImageFilename() string {
	filename := filepath.Join(b.context.ImageConfigDir, b.context.Definition.GetImage().OutputImageName)
	return filename
}

func (b *Builder) generateBaseImageFilename() string {
	filename := filepath.Join(b.context.ImageConfigDir, "base-images", b.context.Definition.GetImage().BaseImage)
	return filename
}

func (b *Builder) deleteExistingOutputImage() error {
	outputFilename := b.generateOutputImageFilename()
	err := os.Remove(outputFilename)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error deleting file %s: %w", outputFilename, err)
	}
	return nil
}
