# DogeIcon

This directory contains a small icon compression library for use in the browser.

## Building

The build product is `dogeicon.js` which is built from `dogeicon.ts`.
Run `npm run build` to rebuild the javascript module.

Building requires Node.js and npm.
There are no runtime dependencies.

## Usage

Here is an example showing how to compress canvas contents:

```
<script src="dogeicon.js"></script>
<canvas id="c" width="48" height="48"></canvas>
<script>
    var c = document.getElementById('c');
    var g = c.getContext("2d", { colorSpace: "srgb" });
    var rgb = g.getImageData(0, 0, 48, 48); // RGBA 9216 bytes
    var options = 24;
    var compressed = DogeIcon.compress(rgb.data, 4, options);
</script>
```

Compress takes the following arguments:

    * RGB or RGBA byte-array
    * Number of components (3 or 4)
    * Options

Options is a bit-field:

    * 1 = linear smoothing (vs pixellated)
    * 4 = use specified smoothing (vs automatic)
    * 8,16 = chroma sampling (0=top-left,8=average,16=horizontal,24=weighted)

## Example

There is an example application in the `web_test` directory.
Run an http server in this directory and access `http://localhost:8000/web_test/upload.html`

Note that the upload script imports `../dogeicon.js` from this directory.
