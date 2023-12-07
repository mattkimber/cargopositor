package compositor

import (
	"encoding/json"
	"fmt"
	"github.com/mattkimber/gandalf/geometry"
	"github.com/mattkimber/gandalf/magica"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Batch struct {
	Files      []string    `json:"files"`
	Operations []Operation `json:"operations"`
}

type BoundingVolume struct {
	Min geometry.Point `json:"min"`
	Max geometry.Point `json:"max"`
}

type Operation struct {
	Name              string          `json:"name"`
	Type              string          `json:"type"`
	File              string          `json:"file"`
	InputColourRamp   string          `json:"input_ramp"`
	OutputColourRamp  string          `json:"output_ramp"`
	InputColourRamps  []string        `json:"input_ramps"`
	OutputColourRamps []string        `json:"output_ramps"`
	N                 int             `json:"n"`
	XSteps            float64         `json:"x_steps"`
	ZSteps            int             `json:"z_steps"`
	Angle             float64         `json:"angle"`
	XOffset           int             `json:"x_offset"`
	YOffset           int             `json:"y_offset"`
	IgnoreMask        bool            `json:"ignore_mask"`
	Truncate          bool            `json:"truncate"`
	MaskOriginal      bool            `json:"mask_original"`
	FlipX             bool            `json:"flip_x"`
	MaskNew           bool            `json:"mask_new"`
	Scale             geometry.PointF `json:"scale"`
	BoundingVolume    BoundingVolume  `json:"bounding_volume"`
	Overwrite         bool            `json:"overwrite"`
	Layers            []int           `json:"layers"`
}

func FromJson(handle io.Reader) (b Batch, err error) {
	data, err := ioutil.ReadAll(handle)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &b)
	return
}

func FromFile(filename string) (b Batch, err error) {
	handle, err := os.Open(filename)
	if err != nil {
		return
	}

	b, err = FromJson(handle)
	if err != nil {
		handle.Close()
		return
	}

	err = handle.Close()
	return
}

func getOutputFileName(directory, filename, suffix string) string {
	// deal with windows paths
	filename = strings.Replace(filename, "\\", "/", -1)
	if len(directory) > 0 {
		filename = directory + path.Base(filename)
	}

	lastExtension := strings.LastIndex(filename, ".")
	if lastExtension != -1 {
		return filename[:lastExtension] + suffix + ".vox"
	}

	return filename + suffix + ".vox"
}

func saveFile(v *magica.VoxelObject, filename string) error {
	handle, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create output file %s: %v", filename, err)
	}

	err = v.Save(handle)
	if err != nil {
		handle.Close()
		return fmt.Errorf("could not open output file: %v", err)
	}

	err = handle.Close()
	return err
}

func (b *Batch) Run(outputDirectory, voxelDirectory string) (err error) {
	if len(voxelDirectory) > 0 && !strings.HasSuffix(voxelDirectory, "/") {
		voxelDirectory = voxelDirectory + "/"
	}

	if len(outputDirectory) > 0 && !strings.HasSuffix(outputDirectory, "/") {
		outputDirectory = outputDirectory + "/"
	}

	// Start with at least the length of files, as we know we have this many
	expandedFiles := make([]string, 0, len(b.Files))

	// Expand file paths
	for _, fileSpec := range b.Files {
		files, err := filepath.Glob(voxelDirectory + fileSpec)
		if err != nil {
			return err
		}
		for _, file := range files {
			expandedFiles = append(expandedFiles, file)
		}
	}

	for _, f := range expandedFiles {
		var input magica.VoxelObject

		for _, op := range b.Operations {

			if len(op.InputColourRamps) == 0 || len(op.InputColourRamps) != len(op.OutputColourRamps) {
				op.InputColourRamps = []string{op.InputColourRamp}
				op.OutputColourRamps = []string{op.OutputColourRamp}
			}

			outputFileName := getOutputFileName(outputDirectory, f, op.Name)

			newer, err := inputFileIsNewerThanOutput(f, voxelDirectory, op.File, outputFileName)
			if err != nil {
				return fmt.Errorf("could not stat input and/or output files: %w", err)
			}

			if !newer {
				continue
			}

			if (input.Size.X == 0 && input.Size.Y == 0 && input.Size.Z == 0) || len(op.Layers) > 0 {
				input, err = magica.FromFileWithLayers(f, op.Layers)
				if err != nil {
					return fmt.Errorf("could not open input file %s: %v", f, err)
				}
			}

			switch op.Type {
			case "identity":
				output := Identity(input)
				if err := saveFile(&output, outputFileName); err != nil {
					return err
				}
			case "produce_empty":
				output := ProduceEmpty(input, op.InputColourRamps, op.OutputColourRamps)
				if err := saveFile(&output, outputFileName); err != nil {
					return err
				}
			case "scale":
				src, err := magica.FromFile(voxelDirectory + op.File)
				if err != nil {
					return fmt.Errorf("error opening voxel file %s: %v", voxelDirectory+op.File, err)
				}
				output := AddScaled(input, src, op.InputColourRamps, op.OutputColourRamps, op.Scale, op.Overwrite, op.IgnoreMask, op.MaskOriginal, op.MaskNew)
				if err := saveFile(&output, outputFileName); err != nil {
					return err
				}
			case "repeat":
				src, err := magica.FromFile(voxelDirectory + op.File)
				if err != nil {
					return fmt.Errorf("error opening voxel file %s: %v", voxelDirectory+op.File, err)
				}
				output := AddRepeated(input, src, op.N, op.InputColourRamps, op.OutputColourRamps, op.Overwrite, op.IgnoreMask, op.Truncate, op.MaskOriginal, op.MaskNew, op.FlipX)
				if err := saveFile(&output, outputFileName); err != nil {
					return err
				}
			case "stairstep":
				output := Stairstep(input, op.XSteps, op.ZSteps)
				if err := saveFile(&output, outputFileName); err != nil {
					return err
				}
			case "rotate":
				output := RotateAndTile(input, op.Angle, op.XOffset, op.YOffset, op.Scale, op.BoundingVolume)
				if err := saveFile(&output, outputFileName); err != nil {
					return err
				}
			case "rotate_y":
				output := RotateY(input, op.Angle)
				if err := saveFile(&output, outputFileName); err != nil {
					return err
				}
			case "rotate_z":
				output := RotateZ(input, op.Angle)
				if err := saveFile(&output, outputFileName); err != nil {
					return err
				}
			case "remove":
				src, err := magica.FromFile(voxelDirectory + op.File)
				if err != nil {
					return fmt.Errorf("error opening voxel file %s: %v", voxelDirectory+op.File, err)
				}
				output := Remove(input, src, 0)
				if err := saveFile(&output, outputFileName); err != nil {
					return err
				}
			case "clip":
				src, err := magica.FromFile(voxelDirectory + op.File)
				if err != nil {
					return fmt.Errorf("error opening voxel file %s: %v", voxelDirectory+op.File, err)
				}
				output := Remove(input, src, 255)
				if err := saveFile(&output, outputFileName); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unkown operation %s", op.Type)
			}
		}
	}

	return
}

func inputFileIsNewerThanOutput(input, voxelDir, opfile, output string) (bool, error) {
	in, err := os.Stat(input)
	if err != nil {
		return false, err
	}

	out, err := os.Stat(output)
	if err != nil {
		// By default if we can't stat the output file, replace it
		return true, nil
	}

	if in.ModTime().After(out.ModTime()) {
		return true, nil
	}

	if opfile != "" {
		in, err := os.Stat(voxelDir + opfile)
		if err != nil {
			return false, err
		}

		if in.ModTime().After(out.ModTime()) {
			return true, nil
		}
	}

	return false, nil
}
