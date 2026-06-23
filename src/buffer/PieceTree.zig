//! A Piece table implementation that instead of a list of pieces uses a treap of pieces.
//! This allows for effecient modifications and inserts without needing to scan through
//! a list. I'm using a treap because I would rather punch myself in the face than
//! implement a red-black tree.
//!
//! One thing to note in this file I refer to "buffer" frequently. This isn't a buffer in
//! the sense of an editor. These are append-only buffers that store contents for the
//! editor buffer.
const PieceTree = @This();

const std = @import("std");

const PieceTreeError = error{
    ContentMissing,
    DuplicateStart,
    OutOfMemory,
};

/// Piece tables have an "original" buffer and an "add" buffer.
/// original - The entire text on first load goes here.
/// add - Any modifications made to the original buffer go here.
const BufferType = enum { original, add };

/// Instead of storing just a start and length in the piece we want to
/// store positions in the buffer. This lets us split pieces without needing
/// to re-read the content.
// const BufferPos = struct {
//     line: usize,
//     column: usize,
// };

/// A node in the piece tree.
const PieceNode = struct {
    buffer_type: BufferType,
    left: ?*PieceNode = null,
    right: ?*PieceNode = null,

    start: usize,
    len: usize,
    // start: BufferPos,
    // end: BufferPos,

    // piece-local length and newline counts
    // len: usize,
    // newline_count: usize,

    // // left subtree lengths and newline counts to make getting pieces by line
    // // and by offset fast.
    // left_subtree_len: usize,
    // left_subtree_newline_count: usize,
};

/// Main tree structure that holds all of the pieces.
const Treap = struct {
    root: *PieceNode,
    alloc: std.mem.Allocator,
    prng: std.Random.Xoshiro256,

    /// Initializes the treap with a root node pointing at the original buffer.
    fn init(alloc: std.mem.Allocator, prng: std.Random.Xoshiro256, initial_content_len: usize) !Treap {
        const piece_node = try alloc.create(PieceNode);
        piece_node.* = .{ .buffer_type = .original, .start = 0, .len = initial_content_len };

        return .{ .alloc = alloc, .prng = prng, .root = piece_node };
    }

    /// Deinit the treap. Clears out all nodes.
    fn deinit(self: *Treap) void {
        self.destroy_nodes(self.root);
    }

    // recursively destroy all nodes in the tree
    fn destroy_nodes(self: *Treap, tree_node: ?*PieceNode) void {
        if (tree_node) |node| {
            self.destroy_nodes(node.left);
            self.destroy_nodes(node.right);
            self.alloc.destroy(node);
        }
    }

    // not working yet, but I think want to do content first so testing is easier
    // /// Inserts a node into the tree.
    // fn insert(tree_node: *PieceNode, insert_node: *PieceNode) PieceTreeError!void {
    //     if (tree_node.buffer_type == insert_node.buffer_type and tree_node.start == insert_node.start) {
    //         return PieceTreeError.DuplicateStart;
    //     } else if (insert_node.start < tree_node.start) {
    //         if (tree_node.left) |left| {
    //             try Treap.insert(left, insert_node);
    //         } else {
    //             tree_node.left = insert_node;
    //         }
    //     } else if (insert_node.start > tree_node.start) {
    //         if (tree_node.right) |right| {
    //             try Treap.insert(right, insert_node);
    //         } else {
    //             tree_node.right = insert_node;
    //         }
    //     }
    // }
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

/// All offsets of line starts in the original buffer.
original_line_starts: std.ArrayList(usize),

/// Append only buffer that holds all additions.
add_buffer: std.ArrayList(u8),

/// All offsets of line starts in the add buffer.
add_line_starts: std.ArrayList(usize),

/// initialize the Piece Tree.
/// The original buffer is generated from the initial contents.
pub fn init(alloc: std.mem.Allocator, prng: std.Random.Xoshiro256, initial_contents: []const u8) !PieceTree {
    var original_buffer: std.ArrayList(u8) = .empty;
    try original_buffer.appendSlice(alloc, initial_contents);

    // find the line starts
    var original_line_starts: std.ArrayList(usize) = .empty;
    try original_line_starts.append(alloc, 0);
    for (initial_contents, 0..) |byte, i| {
        if (byte == '\n') {
            try original_line_starts.append(alloc, i + 1);
        }
    }

    const add_buffer: std.ArrayList(u8) = .empty;
    var add_line_starts: std.ArrayList(usize) = .empty;
    try add_line_starts.append(alloc, 0);

    const treap = try Treap.init(alloc, prng, initial_contents.len);
    return .{
        .alloc = alloc,
        .prng = prng,
        .original_buffer = original_buffer,
        .original_line_starts = original_line_starts,
        .add_buffer = add_buffer,
        .add_line_starts = add_line_starts,
        .pieces = treap,
    };
}

/// Deinit the Piece Tree. It will clear out the buffers and the pieces.
pub fn deinit(self: *PieceTree) void {
    self.original_buffer.deinit(self.alloc);
    self.original_line_starts.deinit(self.alloc);
    self.add_buffer.deinit(self.alloc);
    self.add_line_starts.deinit(self.alloc);
    self.pieces.deinit();
}

/// insert into the piece tree at the given offset.
// pub fn insert(self: *PieceTree, offset: usize, contents: []const u8) PieceTreeError!void {
//     if (contents.len < 1) {
//         return PieceTreeError.ContentMissing;
//     }

//     const add_offset: usize = self.add_buffer.items.len;
//     try self.add_buffer.appendSlice(self.alloc, contents);

//     // inserting at the beginning of the document is a special case, just prepend the node
//     if (offset == 0) {
//         const piece_node = try self.alloc.create(PieceNode);
//         errdefer self.alloc.destroy(piece_node);

//         piece_node.* = .{ .buffer_type = .add, .start = add_offset, .len = contents.len };
//         try Treap.insert(self.pieces.root, piece_node);
//     }
// }

/// Returns the contents for an individual node.
fn node_contents(self: *PieceTree, tree_node: *PieceNode) []u8 {
    const buffer = switch (tree_node.buffer_type) {
        .original => &self.original_buffer,
        .add => &self.add_buffer,
    };

    return buffer.items[tree_node.start .. tree_node.start + tree_node.len];
}

/// Returns all contents of the document. Recursively walk all pieces to assemble the final contents.
pub fn contents(self: *PieceTree, alloc: std.mem.Allocator) ![]u8 {
    var collect_list: std.ArrayList(u8) = .empty;
    errdefer collect_list.deinit(alloc);

    try self.collect_contents(alloc, self.pieces.root, &collect_list);
    return collect_list.toOwnedSlice(alloc);
}

fn collect_contents(self: *PieceTree, alloc: std.mem.Allocator, tree_node: ?*PieceNode, list: *std.ArrayList(u8)) !void {
    if (tree_node) |node| {
        try self.collect_contents(alloc, node.left, list);
        try list.appendSlice(alloc, self.node_contents(node));
        try self.collect_contents(alloc, node.right, list);
    }
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

test "PieceTree init sets the correct line starts for the original buffer" {
    const alloc = std.testing.allocator;
    const prng = std.Random.DefaultPrng.init(0);

    var p = try PieceTree.init(alloc, prng, "one\ntwo\nthree\n");
    defer p.deinit();

    try std.testing.expectEqualSlices(usize, &[_]usize{ 0, 4, 8, 14 }, p.original_line_starts.items);
    try std.testing.expectEqualSlices(usize, &[_]usize{0}, p.add_line_starts.items);
}

test "PieceTree node_contents returns the contents of a node" {
    const alloc = std.testing.allocator;
    const prng = std.Random.DefaultPrng.init(0);

    var p = try PieceTree.init(alloc, prng, "hello\n");
    defer p.deinit();

    try std.testing.expectEqualStrings("hello\n", p.node_contents(p.pieces.root));
}

test "PieceTree contents will get all contents of the tree" {
    const alloc = std.testing.allocator;
    const prng = std.Random.DefaultPrng.init(0);

    var p = try PieceTree.init(alloc, prng, "two\nthree\n");
    defer p.deinit();

    try p.add_buffer.appendSlice(alloc, "one\n");
    try p.add_buffer.appendSlice(alloc, "four\n");

    // left node
    const left = try alloc.create(PieceNode);
    left.* = .{ .buffer_type = .add, .start = 0, .len = 4 };
    p.pieces.root.left = left;

    // right node
    const right = try alloc.create(PieceNode);
    right.* = .{ .buffer_type = .add, .start = 4, .len = 5 };
    p.pieces.root.right = right;

    const content = try p.contents(alloc);
    defer alloc.free(content);

    try std.testing.expectEqualStrings("one\ntwo\nthree\nfour\n", content);
}

// test "PieceTree insert will insert at beginning of the content" {
//     const alloc = std.testing.allocator;
//     const prng = std.Random.DefaultPrng.init(0);

//     var p = try PieceTree.init(alloc, prng, "one\ntwo\nthree\n");
//     defer p.deinit();

//     try p.insert(0, "hello");
//     try std.testing.expectEqualStrings("hello", p.add_buffer.items);
//     // try std.testing.expectEqual(5, p.pieces.root.len);
//     // try std.testing.expectEqual(0, p.pieces.root.start);

// }
