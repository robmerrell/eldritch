# Eldritch

Personalized text editor heavily inspired by Kakoune, Helix, and Nano Emacs.

I wanted to have a project that I could use daily and not use AI while working in it because there is value, learning, and satisfaction in the struggle. The intention for this project isn't for anybody else to find useful so some things will be noticeably missing like an extension language. I'll build it directly into the code without configurability.

## TODO
### Selections
- [ ] Collapsed directional movements
- [ ] Shift without breaking anchors
- [ ] Moving back and forth from long lines to shorter lines
- [ ] Shift by counts greater than 1
- [ ] Select in
  - [ ] Word
  - [ ] Symbols ```[]()""''``<>{}```
  - [ ] Indentation
  - [ ] File
- [ ] Go to
  - [ ] Beginning buffer
  - [ ] End buffer
  - [ ] Beginng line
  - [ ] End line
- [ ] add surrounds
- [ ] change surrounds
- [ ] Go to matching pair
- [ ] Split selections
- [ ] Swap anchor and head

### Buffers
- [ ] Save/Load
  - [ ] Save as
- [ ] Watch for external changes
- [ ] Change contents slice to a rope or other data structure
- [ ] Prev buffer/next buffer by split
- [ ] Mark as dirty
- [ ] Search
- [ ] Multiple selections
- [ ] Input states
- [ ] Input
  - [ ] Insert
  - [ ] Delete
  - [ ] Paste
- [ ] Undo/Redo
- [ ] Autopairs
- [ ] Language support
  - [ ] Indentation
  - [ ] Hooks
- [ ] Copy to clipboard
- [ ] Shackle size, position, etc.

### Modes
- [ ] Insert
- [ ] Command
- [ ] Shell Command

### UI
- [ ] Draw the current buffer
- [ ] Draw selections
- [ ] Modeline
  - [ ] Current mode
  - [ ] Dirty buffer
  - [ ] File offset
  - [ ] Language type
  - [ ] nyan-mode, but FF?
- [ ] Color schemes
  - https://www.beesmygod.com/significant-colors-in-bloodborne/
- [ ] Splits
  - [ ] Vertical splits
  - [ ] Horizontal splits
- [ ] Which key
- [ ] Syntax highlighting
- [ ] Completions
- [ ] Snippets
- [ ] Indentation guides
- [ ] Errors and warnings
- [ ] Git gutters
- [ ] Prompts
- [ ] Eldoc style signatures?
- [ ] Line numbers
- [ ] Soft wrapping

### LSP
- [ ] Multiple servers
- [ ] Completions
- [ ] Symbols
- [ ] Errors
- [ ] Actions
- [ ] Documentation

### Pickers
- [ ] Base component (vertico + marginalia)
- [ ] File picker
  - [ ] Tracked in git
  - [ ] Not tracked in git (all files)
- [ ] Symbol picker
- [ ] Grep
- [ ] Command picker
- [ ] Git changes

### Keys
- [ ] c-g as escape hatch
- [ ] escape to exit insert mode
- [ ] : or alt-x to enter command mode
- [ ] hjkl
- [ ] Prefix with count
- [ ] by word
- [ ] select line
- [ ] select document
- [ ] symbol?
- [ ] jump to word
- [ ] git change hunk
- [ ] till
- [ ] backtill
- [ ] comment
- [ ] redo last uncollapsed selection
- [ ] replace
- [ ] Go to next ] and previous [
  - [ ] problem
  - [ ] buffer
  - [ ] git change hunk
- [ ] Leader
  - [ ] File
  - [ ] Bookmarks
  - [ ] Buffer
  - [ ] Code
  - [ ] LSP
  - [ ] Window
  - [ ] Uncategorized or top level
    - [ ] Format JSON
    - [ ] Open in github
    - [ ] Diagnostics
    - [ ] Hover docs
    - [ ] Show docs in buffer

### System
- [ ] Use as $EDITOR
- [ ] Pipe into
- [ ] Run command and ignore output

### Future Wants
- [ ] Multibuffer for all selections in file
- [ ] Git merge tool
- [ ] local PR Tool
- [ ] File tree
- [ ] Bookmarks
- [ ] Tight zellij integration
- [ ] Server/Client like Kakoune
