# Cargopositor

Cargopositor is a tool that takes these MagicaVoxel files:

![Before.](img/truck_before.png)

And produces these as output:

![After.](img/truck_after.png)

If you create sprites using [GoRender](https://github.com/mattkimber/gorender)
then it will be obvious why this is useful. Cargopositor saves the need to
draw and edit multiple vehicle objects for different cargo graphics, making
it easier and quicker to create sprites and update them.

## Usage

Cargopositor operates on **batches** - JSON files telling it what objects
to load and what operations to perform on them. A typical batch might look
like this:

```json
{
  "files": [
    "example_input.vox"
  ],
  "operations": [
    {
      "type":  "produce_empty"
    }
  ]
}
```

Batches have the following elements:

* `files` - the MagicaVoxel files used as input objects.
* `operations` - the list of operations to perform. Each operation has a mandatory `type` and may also have its own additional fields.

### Input Files

Input .vox files are standard MagicaVoxel objects, with colour **255** used
to indicate areas which can be replaced with the various cargo elements.

### Operations

The following operations are supported:

#### produce_empty

Remove all Cargopositor-specific behaviour voxels from the object. This is
useful for producing the "base" object, so you do not need two files for
"with cargo" and "without cargo" models.

#### scale

Scales the input across the cargo area. This is most useful for bulk cargo
and other cargoes which do not suffer adversely from being stretched in
dimensions.

Sometimes it may not be desirable to scale the object in all dimensions
across the available area, so this can be reduced by adding a `scale`
directive to the operation:

```json
"scale": {
    "x": 0.0,
    "y": 0.5,
    "z": 1.0
}
```

The value determines how much of the original source object's size to
preserve. `1.0` means to preserve completely the original size, and `0.0`
means to use the scaled size, with values between interpolated linearly.

If the source object is larger than the destination and a scaling value
other than `0.0` is used it will be clipped, ultimately turning scale 
into a `repeat` operation with `n = 1` that will copy objects
larger than the destination area.

Supports recolouring.

#### repeat

Repeats the input across the cargo area. This is most useful for crates,
metal coils and other cargo which is in discrete units. Note that the
cargo must be no larger than the destination area.

There is an additional parameter `n` which can be set to non-zero to limit
the number of repeated items.

Supports recolouring.

#### Recolouring

Recolouring is not an operation in itself but is supported by some of
the other operations. It allows you to reassign a ramp of colours from
the input to a new ramp for the output, e.g. to re-colour a pile of
grain to a pile of copper ore.

To set up recolouring, add the following lines to the operation:

```json
"input_ramp": "3,12",
"output_ramp": "72,79"
```

Colour indexes in the input ramp will be linearly interpolated to the output
ramp when these are present.

### Examples

An example JSON file with several operations configured can be found in
the `samples` directory.

### Tips and Tricks

Although any voxels with colour 255 will be cleared from the main voxel
object when compositing, if they are present in the **cargo** they will
be kept.

This allows for the creation of multi-pass setups, where a vehicle chassis
is first composited with various bodies, and each body is then
composited with the appropriate cargo.

![Demo image.](img/multipass.png)

In the above example, a large number of different vehicles could potentially
be produced by changing only the leftmost "base" truck object.