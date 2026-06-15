//! A Piece table implementation that instead of a list of pieces uses a treap of pieces.
//! This allows for effecient modifications and inserts without needing to scan through
//! a list. I'm using a treap because I would rather punch myself in the face than
//! implement a red-black tree.
//!
//! One thing to not in this file I refer to "buffer" frequently. This isn't a buffer in
//! the sense of an editor. These are append-only buffers that store contents for the
//! editor buffer.
const PieceTree = @This();

const std = @import("std");

/// Piece tables have an "original" buffer and an "add" buffer.
/// original - The entire text on first load goes here.
/// add - Any modifications made to the original buffer go here.
const BufferType = enum { original, add };

/// A node in the piece tree.
const PieceNode = struct {
    buffer_type: BufferType,
    start: usize,
    len: usize,

    fn init() PieceNode {}
};

/// Main tree structure that holds all of the pieces.
const Treap = struct {
    root: *PieceNode,
    alloc: std.mem.Allocator,
    prng: std.Random.Xoshiro256,

    /// Initializes the treap with a root node pointing at the original buffer.
    fn init(alloc: std.mem.Allocator, prng: std.Random.Xoshiro256, initial_content_len: usize) !Treap {
        const piece_node = try alloc.create(PieceNode);
        piece_node.* = .{ .buffer_type = .add, .start = 0, .len = initial_content_len };

        return .{ .alloc = alloc, .prng = prng, .root = piece_node };
    }

    /// Deinit the treap. Clears out all nodes.
    fn deinit(self: *Treap) void {
        // delete all nodes
        self.alloc.destroy(self.root);
    }

    // fn insert(self: *Treap, documentOffset: u32, contents: []u8) !void {}
};

/// Treap structure that store the pieces for the piece tree.
pieces: Treap,

/// Allocator used for the the 2 tree append-only buffers and the pieces.
alloc: std.mem.Allocator,

/// Random number generator used for the treap node priorities.
prng: std.Random.Xoshiro256,

/// Buffer that holds the original content. Generally this buffer should never see changes,
/// but there might be some optimizations to be had "reopening" a file so remove pieces
/// from the tree.
original_buffer: std.ArrayList(u8),

/// Append only buffer that holds all additions.
add_buffer: std.ArrayList(u8),

/// initialize the Piece Tree.
///
/// The original buffer is generated from the initial contents.
pub fn init(alloc: std.mem.Allocator, prng: std.Random.Xoshiro256, initial_contents: []const u8) !PieceTree {
    var original_buffer: std.ArrayList(u8) = .empty;
    try original_buffer.appendSlice(alloc, initial_contents);

    const add_buffer: std.ArrayList(u8) = .empty;

    const treap = try Treap.init(alloc, prng, initial_contents.len);
    return .{ .alloc = alloc, .prng = prng, .add_buffer = add_buffer, .original_buffer = original_buffer, .pieces = treap };
}

/// Deinit the Piece Tree. It will clear out the buffers and the pieces.
pub fn deinit(self: *PieceTree) void {
    self.original_buffer.deinit(self.alloc);
    self.add_buffer.deinit(self.alloc);
    self.pieces.deinit();
}

test "PieceTree init puts contents into original buffer and creates a tree for the pieces" {
    const alloc = std.testing.allocator;
    const prng = std.Random.DefaultPrng.init(0);

    var p = try PieceTree.init(alloc, prng, "hello");
    defer p.deinit();

    try std.testing.expectEqualStrings("hello", p.original_buffer.items);
    try std.testing.expectEqual(5, p.pieces.root.len);
    try std.testing.expectEqual(0, p.pieces.root.start);
}
