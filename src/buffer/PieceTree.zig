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
    parent: ?*PieceNode = null,

    start: usize,
    len: usize,

    // caches for faster lookup
    left_subtree_len: usize,

    // start: BufferPos,
    // end: BufferPos,

    // piece-local length and newline counts
    // newline_count: usize,

    // // left subtree lengths and newline counts to make getting pieces by line
    // // and by offset fast.
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
        piece_node.* = .{
            .buffer_type = .original,
            .start = 0,
            .len = initial_content_len,
            .left_subtree_len = 0,
        };

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

/// The last piece and end that was inserted into. We use this to check if we can just keep adding to the piece.
last_insert_node: ?*PieceNode = null,
last_insert_end: usize = 0,

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
pub fn insert(self: *PieceTree, offset: usize, add_contents: []const u8) PieceTreeError!void {
    if (add_contents.len < 1) {
        return PieceTreeError.ContentMissing;
    }

    const add_offset: usize = self.add_buffer.items.len;
    try self.add_buffer.appendSlice(self.alloc, add_contents);

    // try to grow the last written node
    if (self.last_insert_node) |node| {
        if (node.buffer_type == .add and
            node.start + node.len == add_offset and
            offset == self.last_insert_end)
        {
            node.len += add_contents.len;
            update_caches(node.parent, node, add_contents.len);
            self.last_insert_end += add_contents.len;
            return;
        }
    }

    const node_loc = node_at_offset(self.pieces.root, offset);
    if (node_loc.node) |node| {
        if (node_loc.local_offset == 0) {
            // create a new piece and place on left side of parent and relink left child further down
            const new_node = try self.alloc.create(PieceNode);
            new_node.* = .{
                .buffer_type = .add,
                .start = add_offset,
                .len = add_contents.len,
                .left_subtree_len = node.left_subtree_len,
                .parent = node,
                .left = node.left,
            };
            node.left = new_node;

            if (new_node.left) |left_node| {
                left_node.parent = new_node;
            }

            update_caches(node, new_node, add_contents.len);
            self.last_insert_node = new_node;
            self.last_insert_end = offset + add_contents.len;
        } else {
            // middle of a piece - Create a new left piece, Create a new right, and modify current piece to point at new content

            // new sides
            const new_left = try self.alloc.create(PieceNode);
            new_left.* = .{
                .buffer_type = node.buffer_type,
                .start = node.start,
                .len = node_loc.local_offset,
                .left_subtree_len = node.left_subtree_len,
                .parent = node,
                .left = node.left,
                .right = null,
            };

            const new_right = try self.alloc.create(PieceNode);
            new_right.* = .{
                .buffer_type = node.buffer_type,
                .start = node.start + node_loc.local_offset,
                .len = node.len - node_loc.local_offset,
                .left_subtree_len = 0,
                .parent = node,
                .left = null,
                .right = node.right,
            };

            // modify the current
            node.buffer_type = .add;
            node.left = new_left;
            node.right = new_right;
            node.start = add_offset;
            node.len = add_contents.len;
            node.left_subtree_len += new_left.len;

            // modify the subtrees that moved around for the new nodes
            if (new_left.left) |left_node| {
                left_node.parent = new_left;
            }
            if (new_right.right) |right_node| {
                right_node.parent = new_right;
            }

            update_caches(node.parent, node, add_contents.len);
            self.last_insert_node = node;
            self.last_insert_end = offset + add_contents.len;
        }
    } else {
        // TODO: do something when out of bounds?
        @panic("Out of bounds insert");
    }
}

const NodeLocation = struct { node: ?*PieceNode, local_offset: usize };

/// Returns the node that contains the given offset.
fn node_at_offset(tree_node: ?*PieceNode, offset: usize) NodeLocation {
    if (tree_node) |node| {
        // keep going left
        if (offset < node.left_subtree_len) {
            return node_at_offset(node.left, offset);
        }

        // found on the node
        const local = offset - node.left_subtree_len;
        if (local < node.len) {
            return NodeLocation{ .node = node, .local_offset = local };
        }

        // go right
        return node_at_offset(node.right, offset - node.left_subtree_len - node.len);
    }

    return NodeLocation{ .node = null, .local_offset = 0 };
}

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

/// walk up the tree and update all node size caches
fn update_caches(tree_node: ?*PieceNode, from_node: *PieceNode, delta: usize) void {
    if (tree_node) |node| {
        if (from_node == node.left) {
            node.left_subtree_len += delta;
        }

        update_caches(node.parent, node, delta);
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

// Generates a realistic editing scenario where we have a 7 lines in a document "one\ntwo\n" etc.
// Nodes should be in order. Hopefully this can survive all the changes I'm making to the tree...
// I'm not sure if zig has a more clever way of building fixtures, but this will do.
// The tree structure should look like:
//                  4
//                 / \
//                2   6
//               /\   /\
//              1 3  5 7
fn piece_tree_fixture(alloc: std.mem.Allocator) !PieceTree {
    const prng = std.Random.DefaultPrng.init(0);

    var p = try PieceTree.init(alloc, prng, "four\n");
    try p.add_buffer.appendSlice(alloc, "one\n");
    try p.add_buffer.appendSlice(alloc, "two\n");
    try p.add_buffer.appendSlice(alloc, "three\n");
    try p.add_buffer.appendSlice(alloc, "five\n");
    try p.add_buffer.appendSlice(alloc, "six\n");
    try p.add_buffer.appendSlice(alloc, "seven\n");

    p.pieces.root.left_subtree_len = 14;

    // one|two|three|five|six|seven| -- add buffer
    // one|two|three|four|five|six|seven| -- contents

    // 2
    const node2 = try alloc.create(PieceNode);
    node2.* = .{ .buffer_type = .add, .start = 4, .len = 4, .left_subtree_len = 4, .parent = p.pieces.root };
    p.pieces.root.left = node2;

    // 1
    const node1 = try alloc.create(PieceNode);
    node1.* = .{ .buffer_type = .add, .start = 0, .len = 4, .left_subtree_len = 0, .parent = node2 };
    p.pieces.root.left.?.left = node1;

    // 3
    const node3 = try alloc.create(PieceNode);
    node3.* = .{ .buffer_type = .add, .start = 8, .len = 6, .left_subtree_len = 0, .parent = node2 };
    p.pieces.root.left.?.right = node3;

    // 6
    const node6 = try alloc.create(PieceNode);
    node6.* = .{ .buffer_type = .add, .start = 19, .len = 4, .left_subtree_len = 5, .parent = p.pieces.root };
    p.pieces.root.right = node6;

    // 5
    const node5 = try alloc.create(PieceNode);
    node5.* = .{ .buffer_type = .add, .start = 14, .len = 5, .left_subtree_len = 0, .parent = node6 };
    p.pieces.root.right.?.left = node5;

    // 7
    const node7 = try alloc.create(PieceNode);
    node7.* = .{ .buffer_type = .add, .start = 23, .len = 6, .left_subtree_len = 0, .parent = node6 };
    p.pieces.root.right.?.right = node7;

    return p;
}

test "PieceTree contents will get all contents of the tree" {
    const alloc = std.testing.allocator;

    var p = try piece_tree_fixture(alloc);
    defer p.deinit();

    const content = try p.contents(alloc);
    defer alloc.free(content);

    try std.testing.expectEqualStrings("one\ntwo\nthree\nfour\nfive\nsix\nseven\n", content);
}

test "PieceTree node_at_offset finds the correct nodes at their offsets" {
    const alloc = std.testing.allocator;

    var p = try piece_tree_fixture(alloc);
    defer p.deinit();

    const content = try p.contents(alloc);
    defer alloc.free(content);

    // out of bounds
    var node_loc = node_at_offset(p.pieces.root, 1000);
    try std.testing.expect(node_loc.node == null);
    try std.testing.expectEqual(0, node_loc.local_offset);

    // one
    node_loc = node_at_offset(p.pieces.root, 1);
    try std.testing.expectEqual(p.pieces.root.left.?.left, node_loc.node);
    try std.testing.expectEqual(1, node_loc.local_offset);

    // two
    node_loc = node_at_offset(p.pieces.root, 4);
    try std.testing.expectEqual(p.pieces.root.left, node_loc.node);
    try std.testing.expectEqual(0, node_loc.local_offset);

    // three
    node_loc = node_at_offset(p.pieces.root, 13);
    try std.testing.expectEqual(p.pieces.root.left.?.right, node_loc.node);
    try std.testing.expectEqual(5, node_loc.local_offset);

    // four
    node_loc = node_at_offset(p.pieces.root, 15);
    try std.testing.expectEqual(p.pieces.root, node_loc.node);
    try std.testing.expectEqual(1, node_loc.local_offset);

    // five
    node_loc = node_at_offset(p.pieces.root, 21);
    try std.testing.expectEqual(p.pieces.root.right.?.left, node_loc.node);
    try std.testing.expectEqual(2, node_loc.local_offset);

    // six
    node_loc = node_at_offset(p.pieces.root, 25);
    try std.testing.expectEqual(p.pieces.root.right, node_loc.node);
    try std.testing.expectEqual(1, node_loc.local_offset);

    // seven
    node_loc = node_at_offset(p.pieces.root, 31);
    try std.testing.expectEqual(p.pieces.root.right.?.right, node_loc.node);
    try std.testing.expectEqual(3, node_loc.local_offset);
}

test "update_caches updates left_subtree_len for all parents" {
    const alloc = std.testing.allocator;

    var p = try piece_tree_fixture(alloc);
    defer p.deinit();

    const content = try p.contents(alloc);
    defer alloc.free(content);

    // left side traversal causes update (update node "1")
    var node = p.pieces.root.left.?.left.?;
    update_caches(node.parent, node, 3);
    try std.testing.expectEqual(7, p.pieces.root.left.?.left_subtree_len);
    try std.testing.expectEqual(17, p.pieces.root.left_subtree_len);

    // right does not until a left side is used again (update node "3")
    node = p.pieces.root.left.?.right.?;
    update_caches(node.parent, node, 3);
    try std.testing.expectEqual(7, p.pieces.root.left.?.left_subtree_len);
    try std.testing.expectEqual(20, p.pieces.root.left_subtree_len);

    // subtree of right updates (update node "5")
    node = p.pieces.root.right.?.left.?;
    update_caches(node.parent, node, 5);
    try std.testing.expectEqual(10, p.pieces.root.right.?.left_subtree_len);
    try std.testing.expectEqual(20, p.pieces.root.left_subtree_len); // still the same from node 3 change

    // all rights do not update
    node = p.pieces.root.right.?.right.?;
    update_caches(node.parent, node, 5);
    try std.testing.expectEqual(10, p.pieces.root.right.?.left_subtree_len);
    try std.testing.expectEqual(20, p.pieces.root.left_subtree_len);
}

test "PieceTree insert can insert at the beginning of a piece" {
    const alloc = std.testing.allocator;

    var p = try piece_tree_fixture(alloc);
    defer p.deinit();

    // splits nodes - adds "2" to the beginning of "two".
    try p.insert(4, "2");
    const content = try p.contents(alloc);
    defer alloc.free(content);
    try std.testing.expectEqualStrings("one\n2two\nthree\nfour\nfive\nsix\nseven\n", content);

    // test the node setup
    const new_node = p.pieces.root.left.?;
    try std.testing.expectEqual(p.pieces.root, new_node.parent);
    try std.testing.expectEqual(5, new_node.left_subtree_len);
    try std.testing.expectEqual(15, new_node.parent.?.left_subtree_len);

    const moved_node = p.pieces.root.left.?.left.?;
    try std.testing.expectEqual(4, moved_node.left_subtree_len);
}

test "PieceTree insert can insert in the middle of a piece" {
    const alloc = std.testing.allocator;

    var p = try piece_tree_fixture(alloc);
    defer p.deinit();

    // splits nodes adds "4" to the middle of "four".
    try p.insert(16, "4");
    const content = try p.contents(alloc);
    defer alloc.free(content);
    try std.testing.expectEqualStrings("one\ntwo\nthree\nfo4ur\nfive\nsix\nseven\n", content);

    // updated node
    const updated_node = p.pieces.root;
    try std.testing.expectEqual(1, updated_node.len);
    try std.testing.expectEqual(16, updated_node.left_subtree_len);

    // new left
    const new_left = updated_node.left.?;
    try std.testing.expectEqual(2, new_left.len);
    try std.testing.expectEqual(14, new_left.left_subtree_len);

    // new right
    const new_right = updated_node.right.?;
    try std.testing.expectEqual(3, new_right.len);
    try std.testing.expectEqual(0, new_right.left_subtree_len);
}

test "PieceTree insert can insert at the end of a piece" {
    const alloc = std.testing.allocator;

    var p = try piece_tree_fixture(alloc);
    defer p.deinit();

    // splits nodes adds "1" to the end of "one".
    try p.insert(3, "1");
    const content = try p.contents(alloc);
    defer alloc.free(content);
    try std.testing.expectEqualStrings("one1\ntwo\nthree\nfour\nfive\nsix\nseven\n", content);

    // nodes
    const parent = p.pieces.root.left.?.left.?;
    try std.testing.expectEqual(1, parent.len);
    try std.testing.expectEqual(3, parent.left_subtree_len);

    const new_left = parent.left.?;
    try std.testing.expectEqual(3, new_left.len);
    try std.testing.expectEqual(0, new_left.left_subtree_len);
}

test "PieceTree insert will continue writing on a piece" {
    const alloc = std.testing.allocator;

    var p = try piece_tree_fixture(alloc);
    defer p.deinit();

    // splits nodes adds "1" to the end of "one".
    try p.insert(3, "1");

    // then keeps inserting onto it
    try p.insert(4, "2345");
    const content = try p.contents(alloc);
    defer alloc.free(content);
    try std.testing.expectEqualStrings("one12345\ntwo\nthree\nfour\nfive\nsix\nseven\n", content);

    // nodes
    const parent = p.pieces.root.left.?.left.?;
    try std.testing.expectEqual(5, parent.len);
    try std.testing.expectEqual(9, parent.parent.?.left_subtree_len);
}

test "PieceTree insert can handle thousands of inserts" {
    // Fuzz testing is broken on Linux at the moment, so this fakes it a bit by creating a reference
    // buffer. We insert into the reference buffer and the piece tree at the same location and compare
    // their strings outputs. They should always match.
    const alloc = std.testing.allocator;

    var p = try piece_tree_fixture(alloc);
    defer p.deinit();

    // match the reference to our piece tree fixture
    var reference: std.ArrayList(u8) = .empty;
    try reference.insertSlice(alloc, 0, "one\ntwo\nthree\nfour\nfive\nsix\nseven\n");
    defer reference.deinit(alloc);

    var prng = std.Random.DefaultPrng.init(0);
    const rand = prng.random();

    for (1..2000) |_| {
        const pos = rand.intRangeLessThan(usize, 0, reference.items.len);
        const content_len = rand.intRangeAtMost(usize, 1, 100);

        // generate random contents to insert
        var insert_contents: std.ArrayList(u8) = .empty;
        defer insert_contents.deinit(alloc);
        try insert_contents.appendNTimes(alloc, @as(u8, 'a'), content_len);

        // insert into both
        try reference.insertSlice(alloc, pos, insert_contents.items);
        try p.insert(pos, insert_contents.items);

        // std.debug.print("insert len: {d}, at: {d}\n", .{ content_len, pos });

        // compare
        const tree_contents = try p.contents(alloc);
        defer alloc.free(tree_contents);
        try std.testing.expectEqualStrings(reference.items, tree_contents);
    }
}
