# Cargopositor

Cargopositor is a tools that takes these MagicaVoxel files:

![Before.](./img/truck_before.png)

And produces these as output:

![After.](./img/truck_after.png)

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