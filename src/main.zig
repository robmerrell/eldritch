const std = @import("std");
const Io = std.Io;

pub fn main(init: std.process.Init) !void {
    std.debug.print("hello", .{});
    _ = init.arena.allocator();
    // const _: std.mem.Allocator = init.arena.allocator();
    // const io = init.io;
}

// I don't think I need this forever, but handy for now I think...I'm probably doing something wrong.
test {
    _ = @import("buffer/PieceTree.zig");
}
