"use strict";
(function(){
    var u = document.getElementById('u');
    var cm = document.getElementById('cm');
    var pre = document.getElementById('pre');
    var mt = document.getElementById('mt');

    var c = document.getElementById('c');
    var d = document.getElementById('d');
    var o = document.getElementById('o');
    var dl = document.getElementById('dl');
    var g = c.getContext("2d", { colorSpace: "srgb" });
    var dg = d.getContext("2d", { colorSpace: "srgb" });
    var og = o.getContext("2d", { colorSpace: "srgb" });
    var dlg = dl.getContext("2d", { colorSpace: "srgb" });
    var dpi = window.devicePixelRatio || 1;

    var size = 48;
    var scale = 4;
    var blend = 24; // 24=weighted 16=horizontal 8=average 0=top-left
    var mode = 1; // 1=linear 0=pixellated -1=auto
    drawmode();

    c.width = size;
    c.height = size;
    var viewsize = size * scale;
    c.style.width = viewsize + 'px';
    c.style.height = viewsize + 'px';
    c.style.imageRendering = 'pixelated';

    d.width = size
    d.height = size
    d.style.width = size + 'px';
    d.style.height = size + 'px';
    d.style.imageRendering = 'pixelated';

    o.width = size
    o.height = size
    o.style.width = size + 'px';
    o.style.height = size + 'px';
    o.style.imageRendering = 'pixelated';

    dl.width = size*2
    dl.height = size
    dl.style.width = size*2 + 'px';
    dl.style.height = size + 'px';
    dl.style.imageRendering = 'pixelated';

    var url = "";
    var pic = null;
    var zoom = 1;
    var w = 0;
    var h = 0;
    var x = 0;
    var y = 0;
    var tx = 0;
    var ty = 0;
    var drag = false;

    redraw();

    u.addEventListener('change', function(ev) {
        if (ev.target.files && ev.target.files.length === 1) {
            if (url) URL.revokeObjectURL(url);
            url = URL.createObjectURL(ev.target.files[0]);
            pre.height = size; pre.src = url; // preview hack (remove)
            var img = document.createElement('img');
            img.onload = () => {
                pic = img;
                //zoom = Math.min(size/img.width, size/img.height); // contain
                zoom = Math.max(size/img.width, size/img.height); // cover
                w = pic.width * zoom;
                h = pic.height * zoom;
                x = (size - w)/2; // centre
                y = (size - h)/2;
                console.log("zoom",zoom);
                redraw();
                compress();
            };
            img.src = url;
        }
    });

    function clamp() {
        w = pic.width * zoom;
        h = pic.height * zoom;
        if (w <= size) {
            x = (size - w)/2; // centre
        } else {
            // x is usually negative (off the left edge)
            if (x < size - w) x = size - w; // stop at right (-x)
            else if (x > 0) x = 0; // stop at left (+x)
        }
        if (h <= size) {
            y = (size - h)/2; // centre
        } else {
            // y is usually negative (off the top edge)
            if (y < size - h) y = size - h; // stop at bottom (-y)
            if (y > 0) y = 0; // stop at top (+y)
        }
    }

    function redraw() {
        g.fillStyle = "#fff";
        g.fillRect(0, 0, size, size);
        if (pic) {
            clamp();
            console.log("draw", x, y, w, h);
            // g.imageSmoothingDisabled = true;
            g.drawImage(pic, x, y, w, h); // compressed zoom
            dg.drawImage(pic, x, y, w, h); // compressed actual size
            og.drawImage(pic, x, y, w, h); // original
        }
        g.fillStyle = "rgba(0,0,0,0.6)";
        g.fillRect(0, 0, size, 0);
        g.fillRect(0, size, size, 0);
        g.fillRect(0, 0, 0, size);
        g.fillRect(size, 0, 0, size);
    }

    c.addEventListener('mousedown', function(ev) {
        ev.preventDefault();
        // screen to doc space
        tx = (ev.clientX * zoom * scale) - x;
        ty = (ev.clientY * zoom * scale) - y;
        console.log(zoom)
        drag = true;
    });

    c.addEventListener('mousemove', function(ev) {
        ev.preventDefault();
        if (!drag) return;
        // screen to doc space
        x = (ev.clientX * zoom * scale) - tx;
        y = (ev.clientY * zoom * scale) - ty;
        clamp();
        window.requestAnimationFrame(compress);
    });

    c.addEventListener('mouseup', function(ev) {
        ev.preventDefault();
        drag = false;
    });

    c.addEventListener('touchcancel', function(ev) {
        ev.preventDefault();
        drag = false;
    });

    function drawmode() {
        var s = (mode==1)?'lin.':'px.';
        if (mode==-1) s = 'auto.';
        switch(blend){
            case 0: s+='TL'; break;
            case 8: s+='Avg'; break;
            case 16: s+='Hz'; break;
            case 24: s+='WAvg'; break;
        }
        mt.innerText = s;
    }

    window.addEventListener('keydown', function(ev) {
        if (ev.key == "+" || ev.key == "=") { zoom = zoom * 1.25; clamp(); compress(); }
        if (ev.key == "-" || ev.key == "_") { zoom = zoom * 0.8; clamp(); compress(); }
        if (ev.key == "r") compress();
        if (ev.key == "c") { blend=24; mode=1; drawmode(); compress(); } // weighted average
        if (ev.key == "1") { blend=0; drawmode(); compress(); }  // top-left chroma sample
        if (ev.key == "2") { blend=8; drawmode(); compress(); } // average 4 chroma
        if (ev.key == "3") { blend=16; drawmode(); compress(); } // average 2 chroma (horizontal) -- best for "aurora"
        if (ev.key == "4") { blend=24; drawmode(); compress(); }  // weighted average
        if (ev.key == "l") { mode = 1; drawmode(); compress(); } // linear
        if (ev.key == "p") { mode = 0; drawmode(); compress(); } // pixellated
        if (ev.key == "o") { mode = -1; drawmode(); compress(); } // auto (min.sad)
        if (ev.key == "ArrowUp") { y -= 1; clamp(); compress(); }
        if (ev.key == "ArrowDown") { y += 1; clamp(); compress(); }
        if (ev.key == "ArrowLeft") { x -= 1; clamp(); compress(); }
        if (ev.key == "ArrowRight") { x += 1; clamp(); compress(); }
    });

    cm.addEventListener('click', function(ev) {
        compress();
    });

    async function compress() {
        if (!pic) return;
        redraw(); // source for getImageData
        var opt = blend;
        if (mode >= 0) opt |= 4 + mode;
        var snap = og.getImageData(0, 0, 48, 48); // RGBA
        dlg.putImageData(snap, size, 0); // uncompressed (right side)
        console.log("compress", blend, mode);
        var res = DogeIcon.compress(snap.data, 4, opt);
        var from = res.res; // result of compression, uncompressed
        if (from.length != 6912) { console.log("Wrong size: "+from.length+" (expected 6912 RGB)"); return; }
        var to = snap.data;
        var p = 0;
        for (var i=0; i<48*48*3; i+=3) {
            to[p] = from[i];
            to[p+1] = from[i+1];
            to[p+2] = from[i+2];
            to[p+3] = 255;
            p += 4;
        }
        g.putImageData(snap, 0, 0); // compressed zoomed
        dg.putImageData(snap, 0, 0); // compressed actual
        dlg.putImageData(snap, 0, 0); // compressed (left side)
    }

})();
