const std = @import("std");
const vaxis = @import("vaxis");
const Event = @import("event.zig").Event;
const Cell = vaxis.Cell;
const TextInput = vaxis.widgets.TextInput;
const border = vaxis.widgets.border;

const App = @This();

/// Allocator used for the TUI.
alloc: std.mem.Allocator,

/// Io instance
io: std.Io,

/// TTY for libvaxis and buffer
buffer: []u8,
tty: vaxis.Tty,

/// Main vaxis instance
vx: vaxis.Vaxis,

pub fn init(io: std.Io, alloc: std.mem.Allocator, env_map: *std.process.Environ.Map) !App {
    // setup a tty and vaxis
    const buffer = try alloc.alloc(u8, 1024);
    const tty = try vaxis.Tty.init(io, buffer);
    const vx = try vaxis.init(io, alloc, env_map, .{});

    return .{
        .alloc = alloc,
        .io = io,
        .buffer = buffer,
        .tty = tty,
        .vx = vx,
    };
}

pub fn deinit(self: *App) void {
    self.vx.deinit(self.alloc, self.tty.writer());
    self.tty.deinit();
    self.alloc.free(self.buffer);
}

pub fn run(self: *App) !void {
    // Start the read loop. This puts the terminal in raw mode and begins
    // reading user input
    var loop: vaxis.Loop(Event) = .init(self.io, &self.tty, &self.vx);
    try loop.start();
    defer loop.stop();

    // Optionally enter the alternate screen
    try self.vx.enterAltScreen(self.tty.writer());

    // We'll adjust the color index every keypress for the border
    var color_idx: u8 = 0;

    // init our text input widget. The text input widget needs an allocator to
    // store the contents of the input
    var text_input = TextInput.init(self.alloc);
    defer text_input.deinit();

    // Sends queries to terminal to detect certain features. This should always
    // be called after entering the alt screen, if you are using the alt screen
    try self.vx.queryTerminal(self.tty.writer(), .fromSeconds(1));

    while (true) {
        // nextEvent blocks until an event is in the queue
        const event = try loop.nextEvent();
        // exhaustive switching ftw. Vaxis will send events if your Event enum
        // has the fields for those events (ie "key_press", "winsize")
        switch (event) {
            .key_press => |key| {
                color_idx = switch (color_idx) {
                    255 => 0,
                    else => color_idx + 1,
                };
                if (key.matches('c', .{ .ctrl = true })) {
                    break;
                } else if (key.matches('l', .{ .ctrl = true })) {
                    self.vx.queueRefresh();
                } else {
                    try text_input.update(.{ .key_press = key });
                }
            },

            // winsize events are sent to the application to ensure that all
            // resizes occur in the main thread. This lets us avoid expensive
            // locks on the screen. All applications must handle this event
            // unless they aren't using a screen (IE only detecting features)
            //
            // The allocations are because we keep a copy of each cell to
            // optimize renders. When resize is called, we allocated two slices:
            // one for the screen, and one for our buffered screen. Each cell in
            // the buffered screen contains an ArrayList(u8) to be able to store
            // the grapheme for that cell. Each cell is initialized with a size
            // of 1, which is sufficient for all of ASCII. Anything requiring
            // more than one byte will incur an allocation on the first render
            // after it is drawn. Thereafter, it will not allocate unless the
            // screen is resized
            .winsize => |ws| try self.vx.resize(self.alloc, self.tty.writer(), ws),
            else => {},
        }

        // vx.window() returns the root window. This window is the size of the
        // terminal and can spawn child windows as logical areas. Child windows
        // cannot draw outside of their bounds
        const win = self.vx.window();

        // Clear the entire space because we are drawing in immediate mode.
        // vaxis double buffers the screen. This new frame will be compared to
        // the old and only updated cells will be drawn
        win.clear();

        // Create a style
        const style: vaxis.Style = .{
            .fg = .{ .index = color_idx },
        };

        // Create a bordered child window
        const child = win.child(.{
            .x_off = win.width / 2 - 20,
            .y_off = win.height / 2 - 3,
            .width = 40,
            .height = 3,
            .border = .{
                .where = .all,
                .style = style,
            },
        });

        // Draw the text_input in the child window
        text_input.draw(child);

        // Render the screen. Using a buffered writer will offer much better
        // performance, but is not required
        try self.vx.render(self.tty.writer());
    }
}
