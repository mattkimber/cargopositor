package compositor

import (
	"encoding/json"
	"fmt"
	"geometry"
	"io"
	"io/ioutil"
	"magica"
	"os"
	"path"
	"strings"
)

type Batch struct {
	Files      []string    `json:"files"`
	Operations []Operation `json:"operations"`
}

type Operation struct {
	Name             string          `json:"name"`
	Type             string          `json:"type"`
	File             string          `json:"file"`
	InputColourRamp  string          `json:"input_ramp"`
	OutputColourRamp string          `json:"output_ramp"`
	N                int             `json:"n"`
	XSteps           float64         `json:"x_steps"`
	ZSteps           int             `json:"z_steps"`
	Angle            float64         `json:"angle"`
	XOffset          int             `json:"x_offset"`
	YOffset          int             `json:"y_offset"`
	IgnoreMask       bool            `json:"ignore_mask"`
	Truncate         bool            `json:"truncate"`
	Scale            geometry.PointF `json:"scale"`
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
		return err
	}

	err = v.Save(handle)
	if err != nil {
		handle.Close()
		return err
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

	for _, f := range b.Files {
		input, err := magica.FromFile(voxelDirectory + f)
		if err != nil {
			return err
		}

		for _, op := range b.Operations {

			switch op.Type {
			case "produce_empty":
				output := ProduceEmpty(input)
				if err := saveFile(&output, getOutputFileName(outputDirectory, f, op.Name)); err != nil {
					return err
				}
			case "scale":
				src, err := magica.FromFile(voxelDirectory + op.File)
				if err != nil {
					return err
				}
				output := AddScaled(input, src, op.InputColourRamp, op.OutputColourRamp, op.Scale, op.IgnoreMask)
				if err := saveFile(&output, getOutputFileName(outputDirectory, f, op.Name)); err != nil {
					return err
				}
			case "repeat":
				src, err := magica.FromFile(voxelDirectory + op.File)
				if err != nil {
					return err
				}
				output := AddRepeated(input, src, op.N, op.InputColourRamp, op.OutputColourRamp, op.IgnoreMask, op.Truncate)
				if err := saveFile(&output, getOutputFileName(outputDirectory, f, op.Name)); err != nil {
					return err
				}
			case "stairstep":
				output := Stairstep(input, op.XSteps, op.ZSteps)
				if err := saveFile(&output, getOutputFileName(outputDirectory, f, op.Name)); err != nil {
					return err
				}
			case "rotate":
				output := Rotate(input, op.Angle, op.XOffset, op.YOffset)
				if err := saveFile(&output, getOutputFileName(outputDirectory, f, op.Name)); err != nil {
					return err
				}
			case "remove":
				src, err := magica.FromFile(voxelDirectory + op.File)
				if err != nil {
					return err
				}
				output := Remove(input, src)
				if err := saveFile(&output, getOutputFileName(outputDirectory, f, op.Name)); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unkown operation %s", op.Type)
			}
		}
	}

	return
}
