const std = @import("std");
const vaxis = @import("vaxis");
const app = @import("tui/app.zig");
const Cell = vaxis.Cell;
const TextInput = vaxis.widgets.TextInput;
const border = vaxis.widgets.border;

pub fn main(init: std.process.Init) !void {
    const io = init.io;
    const alloc = init.gpa;

    var eldApp = try app.init(io, alloc, init.environ_map);
    try eldApp.run();
    defer eldApp.deinit();
}

// I don't think I need this forever, but handy for now I think...I'm probably doing something wrong.
test {
    _ = @import("buffer/PieceTree.zig");
}
