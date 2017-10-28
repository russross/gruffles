jQuery(function ($) {
    var layout = {
        "type": "vsp",
        "panes": [
            {
                "type": "sp",
                "size": 65,
                "panes": [
/*
                    {
                        "type": "output",
                        "size": 45,
                        "id": "combat",
                        "include": "combat"
                    },
*/
                    {
                        "type": "output",
                        "size": 45,
                        "id": "main",
                        "exclude": "combat|social|map"
                    },
                    {
                        "type": "input",
                        "size": 10,
                        "id": "input"
                    }
                ]
            },
            {
                "type": "sp",
                "size": 35,
                "panes": [
/*
                    {
                        "type": "output",
                        "size": 50,
                        "id": "social",
                        "include": "social"
                    },
*/
                    {
                        "type": "map",
                        "size": 50,
                        "id": "map"
                    }
                ]
            }
        ]
    };

    var map, mapText;
    var renderMap = function() {
        if (!map || !mapText)
            return;

        var canvas = map.firstChild;
        var ctx = canvas.getContext('2d');
        ctx.save();

        // convert the string map into a grid of characters
        var grid = [];
        var lines = mapText.split('\n');
        lines.pop();
        for (var i = 0; i < lines.length; i += 1)
            grid.push(lines[i].split(''));
        var size = grid.length;
        var get = function (x, y) {
            if (x < 0 || x >= size || y < 0 || y >= size)
                return ' ';
            return grid[y][x];
        };

        // adjust the canvas to fill the parent container and center it
        canvas.width = map.clientWidth;
        canvas.height = map.clientHeight;
        if (canvas.width > canvas.height)
            ctx.translate((canvas.width - canvas.height) / 2, 0);
        if (canvas.height > canvas.width)
            ctx.translate(0, (canvas.height - canvas.width) / 2);

        // scale the canvas so one character from the input is
        // one unit on the canvas
        var scale = Math.min(canvas.width, canvas.height) / (size+1);
        ctx.scale(scale, scale);

        // translate so a grid is centered
        ctx.translate(0.5, 0.5);

        // draw a grid
        ctx.strokeStyle = 'darkgrey';
        ctx.lineWidth = 0.5 / scale;
        ctx.beginPath();
        for (var i = 0; i <= size; i += 1) {
            ctx.moveTo(i, 0);
            ctx.lineTo(i, size);

            ctx.moveTo(0, i);
            ctx.lineTo(size, i);
        }
        ctx.stroke();

        // clip to the grid boundaries
        ctx.moveTo(0, 0);
        ctx.lineTo(size, 0);
        ctx.lineTo(size, size);
        ctx.lineTo(0, size);
        ctx.lineTo(0, 0);
        ctx.clip();
        ctx.beginPath();

        ctx.strokeStyle = '#c0c0c0';
        ctx.fillStyle = '#c0c0c0';
        ctx.lineWidth = 0.1;
        var ar = 0.4;

        // draw the rooms
        for (var y = -2; y < size+2; y += 4) {
            for (var x = -2; x < size+2; x += 4) {
                ctx.save()

                // translate so all drawing is relative to center of square
                ctx.translate(x + 0.5, y + 0.5);

                // draw a room
                if (get(x-1, y-1) == '╭' && get(x+1, y+1) == '╯') {
                    ctx.beginPath();
                    ctx.moveTo(0.0, -1.0);
                    ctx.quadraticCurveTo( 1.0, -1.0,  1.0,  0.0);
                    ctx.quadraticCurveTo( 1.0,  1.0,  0.0,  1.0);
                    ctx.quadraticCurveTo(-1.0,  1.0, -1.0,  0.0);
                    ctx.quadraticCurveTo(-1.0, -1.0,  0.0, -1.0);
                    if (get(x, y) == '╳')
                        ctx.fill();
                    else
                        ctx.stroke();
                }

                // draw connecting arrows down and to the right
                if (get(x+2, y) == '↔') {
                    ctx.beginPath();
                    ctx.moveTo(1.0, 0.0);
                    ctx.lineTo(3.0, 0.0);
                    ctx.stroke();
                }
                if (get(x, y+2) == '↕') {
                    ctx.beginPath();
                    ctx.moveTo(0.0, 1.0);
                    ctx.lineTo(0.0, 3.0);
                    ctx.stroke();
                }

                // draw single non-connecting arrows in all directions
                var drawSingleArrow = function (rot) {
                    ctx.save();
                    ctx.rotate(rot);
                    ctx.beginPath();
                    ctx.moveTo(1.0, 0.0);
                    ctx.lineTo(1.5, 0.0);
                    ctx.stroke();
                    ctx.beginPath();
                    ctx.moveTo(1.5+ar, 0.0);
                    ctx.lineTo(1.5,    0.0+ar);
                    ctx.lineTo(1.5,    0.0-ar);
                    ctx.lineTo(1.5+ar, 0.0);
                    ctx.fill();
                    ctx.restore();
                };
                if (get(x+1, y) === '→')
                    drawSingleArrow(0);
                if (get(x-1, y) === '←')
                    drawSingleArrow(Math.PI);
                if (get(x, y+1) === '↓')
                    drawSingleArrow(Math.PI * 0.5);
                if (get(x, y-1) === '↑')
                    drawSingleArrow(Math.PI * 1.5);
                if (get(x+1, y-1) == '↗')
                    drawSingleArrow(Math.PI * 1.75);
                if (get(x-1, y+1) == '↙')
                    drawSingleArrow(Math.PI * 0.75);

                var drawDoubleArrow = function (rot) {
                    ctx.save();
                    ctx.rotate(rot);
                    ctx.beginPath();
                    ctx.moveTo(1.0, -0.2);
                    ctx.lineTo(1.5, -0.2);
                    ctx.moveTo(1.0, 0.2);
                    ctx.lineTo(1.5, 0.2);
                    ctx.stroke();
                    ctx.beginPath();
                    ctx.moveTo(1.5+ar, 0.0);
                    ctx.lineTo(1.5,    0.0+ar);
                    ctx.lineTo(1.5,    0.0-ar);
                    ctx.lineTo(1.5+ar, 0.0);
                    ctx.fill();
                    ctx.restore();
                };
                if (get(x+1, y) === '⇒')
                    drawDoubleArrow(0);
                if (get(x-1, y) === '⇐')
                    drawDoubleArrow(Math.PI);
                if (get(x, y+1) === '⇓')
                    drawDoubleArrow(Math.PI * 0.5);
                if (get(x, y-1) === '⇑')
                    drawDoubleArrow(Math.PI * 1.5);
                if (get(x+1, y-1) == '⇗')
                    drawDoubleArrow(Math.PI * 1.75);
                if (get(x-1, y+1) == '⇙')
                    drawDoubleArrow(Math.PI * 0.75);

                var drawDottedArrow = function (rot) {
                    ctx.save();
                    ctx.rotate(rot);
                    ctx.beginPath();
                    ctx.moveTo(1.0, 0.0);
                    ctx.lineTo(1.5, 0.0);
                    ctx.stroke();
                    ctx.restore();
                };
                if (get(x+1, y) === '⇢')
                    drawDottedArrow(0);
                if (get(x-1, y) === '⇠')
                    drawDottedArrow(Math.PI);
                if (get(x, y+1) === '⇣')
                    drawDottedArrow(Math.PI * 0.5);
                if (get(x, y-1) === '⇡')
                    drawDottedArrow(Math.PI * 1.5);
                if (get(x+1, y-1) == '⤴')
                    drawDottedArrow(Math.PI * 1.75);
                if (get(x-1, y+1) == '⤶')
                    drawDottedArrow(Math.PI * 0.75);

                ctx.restore();
            }
        }

        ctx.restore();
    };

    // create the layout
    var outputs = [];
    var input;
    var makeLayout = function makeLayout(div, instructions) {
        if (instructions.id)
            div.id = instructions.id;
        if (instructions.size)
            div.dataset.size = instructions.size;
        else
            div.dataset.size = 1;

        switch (instructions.type) {
        case "vsp":
            var children = [];
            var sizes = [];
            var total = 0;
            for (var i = 0; i < instructions.panes.length; i++) {
                var child = document.createElement('div');
                child.classList.add('split');
                child.classList.add('split-horizontal');
                div.appendChild(child);
                children.push(child);
                makeLayout(child, instructions.panes[i]);
                var size = parseFloat(child.dataset.size);
                sizes.push(size);
                total += size;
            }
            for (var i = 0; i < sizes.length; i++)
                sizes[i] = (sizes[i] / total) * 100.0;
            Split(children, {
                sizes: sizes,
                gutterSize: 8,
                cursor: 'col-resize',
                onDrag: renderMap
            });
            break;

        case "sp":
            var children = [];
            var sizes = [];
            var total = 0;
            for (var i = 0; i < instructions.panes.length; i++) {
                var child = document.createElement('div');
                child.classList.add('split');
                child.classList.add('split-vertical');
                div.appendChild(child);
                children.push(child);
                makeLayout(child, instructions.panes[i]);
                var size = parseFloat(child.dataset.size);
                sizes.push(size);
                total += size;
            }
            for (var i = 0; i < sizes.length; i++)
                sizes[i] = (sizes[i] / total) * 100.0;
            Split(children, {
                direction: 'vertical',
                sizes: sizes,
                gutterSize: 8,
                cursor: 'row-resize',
                onDrag: renderMap
            });
            break;

        case "input":
            div.classList.add('content');
            if (input)
                console.log('cannot have more than one input pane');
            else
                input = div;
            break;

        case "output":
            div.classList.add('content');
            if (instructions.include)
                div.dataset.include = instructions.include;
            if (instructions.exclude)
            div.dataset.exclude = instructions.exclude;
            outputs.push(div);
            break;

        case "map":
            div.classList.add('content');
            if (map)
                console.log('cannot have more than one map pane');
            else
                map = div;
            break;

        default:
            console.log('unknown layout object', instructions);
        }
    };

    makeLayout(document.body, layout);

    // set up the input pane
    if (input) {
        $(input).terminal(function (command) {
            if (command.trim() === '')
                return;
            socket.send(JSON.stringify({"cmd": command}));
        }, {
            greetings: 'Welcome to ' + document.location.hostname + ':' + document.location.port,
            name: 'gruffles input',
            prompt: '] ',
            onBlur: function () { return false; },
            historySize: 500
        });
    } else {
        console.log('must have exactly one input pane');
    }

    // set up the map pane
    if (map) {
        var canvas = document.createElement('canvas');
        canvas.style.width = '100%';
        canvas.style.height = '100%';
        map.style.position = 'relative';
        map.appendChild(canvas);
    }

    // set up the output panes
    for (var i = 0; i < outputs.length; i++) {
        var elt = outputs[i];
        var $elt = $(elt);
        if (elt.dataset.include)
            $elt.data('include', new RegExp(elt.dataset.include));
        if (elt.dataset.exclude)
            $elt.data('exclude', new RegExp(elt.dataset.exclude));
        $elt.terminal(function (command) {
            }, {
                greetings: '',
                name: '',
                prompt: '',
                historySize: 1000
            });
        $elt.freeze(true);
        outputs[i] = $elt;
    }

    var url = 'wss://' + document.location.hostname + ':' + document.location.port + '/server';
    console.log("connecting to " + url);
    var socket = new WebSocket(url);
    socket.onerror = function (event) {
        console.log("websocket error", event);
    };
    socket.onmessage = function (event) {
        // write the message to all applicable output panes
        var data = JSON.parse(event.data);
        if (data.type === 'map') {
            mapText = data.msg;
            renderMap();
            return;
        }
        for (var i = 0; i < outputs.length; i++) {
            var use = true;
            if (outputs[i].data('include'))
                use = outputs[i].data('include').test(data.type);
            if (outputs[i].data('exclude'))
                use = use && !outputs[i].data('exclude').test(data.type);
            if (use)
                outputs[i].echo(data.msg);
        }
    };

});
