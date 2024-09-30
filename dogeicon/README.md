# DogeIcon

This directory contains a small icon compression library for use in the browser.

## Building

The build products are in `dist` and sources in `src`.
Run `npm run build` to rebuild the `dist` javascript modules.

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
    var options = 1+4+24; // linear + weighted average
    var compressed = DogeIcon.compress(rgb.data, 4, options);
</script>
```

Compress takes the following arguments:

    * RGB or RGBA byte-array
    * Number of colour components (3 or 4)
    * Options

Options is a bit-field (or sum) of the following:

    * 0 = pixellated smoothing; or
    * 1 = linear smoothing
    * 4 = turn off auto-detect (use specified smoothing)

And one of these luma-sample options:

    * 0 = use top-left luma sample
    * 8 = average of 4 luma samples
    * 16 = average of 2 horizontal luma samples
    * 24 = weighted average of 4 luma samples (best)

A toggle is useful for pixellated vs linear smoothing; some images
look better with the pixellated method. Use 4+0 or 4+1 to override
the auto-detect, since it often chooses poorly.
Linear smoothing is a good default.

Out of the four luma options, weighted average (24) always works best;
it never looks worse than the other options, and sometimes looks
much better.

## Example

There is an example application in the `web_test` directory.
It works as a local file loaded in a browser.
