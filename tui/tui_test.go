package tui

import (
	"strings"
	"testing"
)

// --- theme tests ---

func TestThemeByName(t *testing.T) {
	for _, name := range []string{"amber", "blue", "green", "purple", "default", "light", "plain"} {
		th, ok := ThemeByName(name)
		if !ok || th == nil {
			t.Errorf("ThemeByName(%q) not found", name)
		}
	}
	_, ok := ThemeByName("nonexistent")
	if ok {
		t.Error("expected false for unknown theme")
	}
}

func TestRegisterTheme(t *testing.T) {
	custom := &Theme{Name: "test-custom", Primary: 0xFF0000}
	RegisterTheme(custom)
	got, ok := ThemeByName("test-custom")
	if !ok || got != custom {
		t.Error("RegisterTheme: theme not retrievable")
	}
}

// --- inputArea tests ---

func TestInputAreaBasic(t *testing.T) {
	a := newInputArea()
	if a.text() != "" {
		t.Error("new inputArea should be empty")
	}
	a.insertRune('h')
	a.insertRune('i')
	if a.text() != "hi" {
		t.Errorf("got %q", a.text())
	}
	if a.charCount() != 2 {
		t.Errorf("charCount: got %d", a.charCount())
	}
}

func TestInputAreaBackspace(t *testing.T) {
	a := newInputArea()
	for _, r := range "hello" {
		a.insertRune(r)
	}
	a.backspace()
	if a.text() != "hell" {
		t.Errorf("got %q", a.text())
	}
	// backspace at start does nothing
	a.home()
	a.backspace()
	if a.text() != "hell" {
		t.Errorf("backspace at start changed text: %q", a.text())
	}
}

func TestInputAreaDeleteForward(t *testing.T) {
	a := newInputArea()
	for _, r := range "abc" {
		a.insertRune(r)
	}
	a.home()
	a.deleteForward()
	if a.text() != "bc" {
		t.Errorf("got %q", a.text())
	}
}

func TestInputAreaMovement(t *testing.T) {
	a := newInputArea()
	for _, r := range "abc" {
		a.insertRune(r)
	}
	a.home()
	if a.col != 0 {
		t.Errorf("home: col=%d", a.col)
	}
	a.end()
	if a.col != 3 {
		t.Errorf("end: col=%d", a.col)
	}
	a.moveLeft()
	if a.col != 2 {
		t.Errorf("moveLeft: col=%d", a.col)
	}
	a.moveRight()
	if a.col != 3 {
		t.Errorf("moveRight: col=%d", a.col)
	}
}

func TestInputAreaCtrlK(t *testing.T) {
	a := newInputArea()
	for _, r := range "hello world" {
		a.insertRune(r)
	}
	// move to position 5
	a.home()
	for i := 0; i < 5; i++ {
		a.moveRight()
	}
	a.ctrlK()
	if a.text() != "hello" {
		t.Errorf("ctrlK: got %q", a.text())
	}
}

func TestInputAreaCtrlU(t *testing.T) {
	a := newInputArea()
	for _, r := range "hello world" {
		a.insertRune(r)
	}
	a.home()
	for i := 0; i < 5; i++ {
		a.moveRight()
	}
	a.ctrlU()
	if a.text() != " world" {
		t.Errorf("ctrlU: got %q", a.text())
	}
}

func TestInputAreaCtrlW(t *testing.T) {
	a := newInputArea()
	for _, r := range "hello world" {
		a.insertRune(r)
	}
	a.ctrlW()
	if a.text() != "hello " {
		t.Errorf("ctrlW: got %q", a.text())
	}
}

func TestInputAreaMultiline(t *testing.T) {
	a := newInputArea()
	for _, r := range "line1" {
		a.insertRune(r)
	}
	a.insertNewline()
	for _, r := range "line2" {
		a.insertRune(r)
	}
	if a.text() != "line1\nline2" {
		t.Errorf("multiline: got %q", a.text())
	}
	if len(a.lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(a.lines))
	}
	// backspace across newline merges lines
	a.home()
	a.backspace()
	if a.text() != "line1line2" {
		t.Errorf("backspace across newline: got %q", a.text())
	}
}

func TestInputAreaHistory(t *testing.T) {
	a := newInputArea()
	a.pushHistory("first")
	a.pushHistory("second")
	// duplicate suppression
	a.pushHistory("second")
	if len(a.history) != 2 {
		t.Errorf("expected 2 history entries, got %d", len(a.history))
	}
	// navigate up
	if !a.historyUp() {
		t.Error("historyUp should return true")
	}
	if a.text() != "second" {
		t.Errorf("historyUp: got %q", a.text())
	}
	if !a.historyUp() {
		t.Error("historyUp should return true")
	}
	if a.text() != "first" {
		t.Errorf("historyUp: got %q", a.text())
	}
	// at oldest — further up returns false
	if a.historyUp() {
		t.Error("historyUp at oldest should return false")
	}
	// navigate back down
	if !a.historyDown() {
		t.Error("historyDown should return true")
	}
	if a.text() != "second" {
		t.Errorf("historyDown: got %q", a.text())
	}
}

func TestInputAreaReset(t *testing.T) {
	a := newInputArea()
	for _, r := range "hello" {
		a.insertRune(r)
	}
	a.reset()
	if a.text() != "" || a.col != 0 || a.row != 0 {
		t.Error("reset did not clear state")
	}
}

// --- palette tests ---

func TestPaletteFilterAndSelect(t *testing.T) {
	cmds := []*Command{
		{Name: "clear", Description: "Clear"},
		{Name: "exit", Description: "Exit"},
		{Name: "theme", Description: "Theme", Args: []string{"amber", "blue"}},
	}
	p := newPalette(cmds)
	p.open("")
	if len(p.filtered) != 3 {
		t.Errorf("expected 3 filtered, got %d", len(p.filtered))
	}
	p.filter("ex")
	if len(p.filtered) != 1 || p.filtered[0].Name != "exit" {
		t.Errorf("filter 'ex': unexpected result")
	}
	if p.selectedCommand() == nil || p.selectedCommand().Name != "exit" {
		t.Error("selectedCommand should be exit")
	}
	p.close()
	if p.active {
		t.Error("palette should be inactive after close")
	}
}

func TestPaletteArgMode(t *testing.T) {
	cmds := []*Command{
		{Name: "theme", Args: []string{"amber", "blue", "green"}},
	}
	p := newPalette(cmds)
	p.open("theme ")
	if !p.argMode {
		t.Error("expected argMode after 'theme '")
	}
	if len(p.argFiltered) != 3 {
		t.Errorf("expected 3 args, got %d", len(p.argFiltered))
	}
	p.filter("theme b")
	if len(p.argFiltered) != 1 || p.argFiltered[0] != "blue" {
		t.Errorf("arg filter 'b': got %v", p.argFiltered)
	}
	if p.selectedArg() != "blue" {
		t.Errorf("selectedArg: got %q", p.selectedArg())
	}
}

func TestPaletteNavigation(t *testing.T) {
	cmds := []*Command{
		{Name: "a"}, {Name: "b"}, {Name: "c"},
	}
	p := newPalette(cmds)
	p.open("")
	p.moveDown(8)
	if p.selected != 1 {
		t.Errorf("moveDown: selected=%d", p.selected)
	}
	p.moveUp()
	if p.selected != 0 {
		t.Errorf("moveUp: selected=%d", p.selected)
	}
	// can't go below 0
	p.moveUp()
	if p.selected != 0 {
		t.Errorf("moveUp at 0: selected=%d", p.selected)
	}
}

// --- outputRegion tests ---

func TestOutputRegionMessages(t *testing.T) {
	o := &outputRegion{
		userLabel:      "You",
		assistantLabel: "Assistant",
		systemLabel:    "System",
	}
	o.AddMessage(RoleUser, "hello")
	o.AddMessage(RoleAssistant, "world")
	if len(o.messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(o.messages))
	}
	o.Clear()
	if len(o.messages) != 0 || o.streaming != nil {
		t.Error("Clear did not reset messages")
	}
}

func TestOutputRegionStreaming(t *testing.T) {
	o := &outputRegion{assistantLabel: "Assistant"}
	o.StartStreaming()
	if o.streaming == nil {
		t.Fatal("streaming should be non-nil")
	}
	o.StreamChunk("hello ")
	o.StreamChunk("world")
	if o.streaming.content != "hello world" {
		t.Errorf("streaming content: %q", o.streaming.content)
	}
	o.StreamComplete()
	if o.streaming != nil {
		t.Error("streaming should be nil after StreamComplete")
	}
	if len(o.messages) != 1 || o.messages[0].content != "hello world" {
		t.Error("streamed message not appended correctly")
	}
}

func TestOutputRegionStreamingAs(t *testing.T) {
	o := &outputRegion{assistantLabel: "Assistant"}
	o.StartStreamingAs("GPT-4o")
	if o.streaming.label != "GPT-4o" {
		t.Errorf("label: %q", o.streaming.label)
	}
	o.StreamComplete()
}

func TestOutputRegionScroll(t *testing.T) {
	o := &outputRegion{}
	o.scrollUp(5)
	if o.scrollOff != 5 {
		t.Errorf("scrollUp: %d", o.scrollOff)
	}
	o.scrollDown(3)
	if o.scrollOff != 2 {
		t.Errorf("scrollDown: %d", o.scrollOff)
	}
	o.scrollDown(10)
	if o.scrollOff != 0 {
		t.Errorf("scrollDown clamp: %d", o.scrollOff)
	}
}

func TestAddMessageAs(t *testing.T) {
	o := &outputRegion{assistantLabel: "Assistant"}
	o.AddMessageAs(RoleAssistant, "Claude", "hi")
	if o.messages[0].label != "Claude" {
		t.Errorf("label: %q", o.messages[0].label)
	}
}

// --- render helpers ---

func TestStripANSI(t *testing.T) {
	s := "\x1b[38;2;255;0;0mhello\x1b[0m"
	got := stripANSI(s)
	if got != "hello" {
		t.Errorf("stripANSI: %q", got)
	}
}

func TestVisibleLen(t *testing.T) {
	s := "\x1b[1mhello\x1b[0m"
	if visibleLen(s) != 5 {
		t.Errorf("visibleLen: %d", visibleLen(s))
	}
}

func TestTruncate(t *testing.T) {
	s := "hello world"
	got := truncate(s, 5)
	if got != "hello" {
		t.Errorf("truncate: %q", got)
	}
	// shorter than limit — unchanged
	got = truncate("hi", 10)
	if got != "hi" {
		t.Errorf("truncate short: %q", got)
	}
}

func TestRenderCodeBlock(t *testing.T) {
	lines := renderCodeBlock("x := 1\n", ThemeAmber, 40)
	if len(lines) < 3 {
		t.Errorf("expected at least 3 lines, got %d", len(lines))
	}
}

func TestRenderMessage(t *testing.T) {
	m := &message{role: RoleAssistant, content: "hello\n\n```go\nfmt.Println()\n```\n"}
	lines := renderMessage(m, ThemeAmber, 80, "You", "Assistant", "System", false)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(stripANSI(joined), "hello") {
		t.Error("rendered message missing content")
	}
}

// --- ANSI helpers ---

func TestANSIHelpers(t *testing.T) {
	if cursorPos(1, 1) != "\x1b[1;1H" {
		t.Errorf("cursorPos: %q", cursorPos(1, 1))
	}
	if clearLine() != "\x1b[2K" {
		t.Errorf("clearLine: %q", clearLine())
	}
	if fg(0) != "" {
		t.Error("fg(0) should be empty")
	}
	if bg(0) != "" {
		t.Error("bg(0) should be empty")
	}
	if !strings.Contains(fg(0xFF0000), "255;0;0") {
		t.Errorf("fg color: %q", fg(0xFF0000))
	}
}

// --- menu tests ---

func TestMenuState(t *testing.T) {
	m := &Menu{
		Title: "Root",
		Items: []*MenuItem{
			{Label: "A"},
			{Label: "B"},
			{Label: "C"},
		},
	}
	ms := newMenuState(m)
	if ms.current().menu.Title != "Root" {
		t.Error("initial menu title wrong")
	}
	ms.moveDown(6)
	if ms.current().selected != 1 {
		t.Errorf("moveDown: selected=%d", ms.current().selected)
	}
	ms.moveUp(6)
	if ms.current().selected != 0 {
		t.Errorf("moveUp: selected=%d", ms.current().selected)
	}

	// push sub-menu
	sub := &Menu{Title: "Sub", Items: []*MenuItem{{Label: "X"}}}
	ms.push(sub)
	if ms.current().menu.Title != "Sub" {
		t.Error("push: wrong title")
	}
	if !ms.pop() {
		t.Error("pop should return true")
	}
	if ms.current().menu.Title != "Root" {
		t.Error("after pop: wrong title")
	}
	// pop at root returns false
	if ms.pop() {
		t.Error("pop at root should return false")
	}
}

func TestMenuRender(t *testing.T) {
	m := &Menu{
		Title: "Test",
		Items: []*MenuItem{
			{Label: "Item 1"},
			{Label: "Item 2"},
		},
	}
	ms := newMenuState(m)
	var buf strings.Builder
	ms.render(&buf, ThemeAmber, 80, 10, 1)
	out := stripANSI(buf.String())
	if !strings.Contains(out, "Test") {
		t.Error("render missing title")
	}
	if !strings.Contains(out, "Item 1") {
		t.Error("render missing item")
	}
}

func TestMenuPromptMode(t *testing.T) {
	item := &MenuItem{Label: "Key", Prompt: "Enter key:"}
	m := &Menu{Title: "Settings", Items: []*MenuItem{item}}
	ms := newMenuState(m)
	lv := ms.current()
	lv.promptItem = item
	lv.promptBuf = []rune("myvalue")

	var buf strings.Builder
	ms.render(&buf, ThemeAmber, 80, 10, 1)
	out := stripANSI(buf.String())
	if !strings.Contains(out, "Enter key:") {
		t.Error("prompt label missing from render")
	}
	if !strings.Contains(out, "myvalue") {
		t.Error("prompt input missing from render")
	}
}

// --- TUI constructor ---

func TestNewDefaults(t *testing.T) {
	tui := New(Config{})
	if tui.theme != ThemeDefault {
		t.Error("default theme should be ThemeDefault")
	}
	if tui.cfg.UserLabel != "You" {
		t.Errorf("UserLabel: %q", tui.cfg.UserLabel)
	}
	if tui.cfg.AssistantLabel != "Assistant" {
		t.Errorf("AssistantLabel: %q", tui.cfg.AssistantLabel)
	}
	if tui.cfg.SystemLabel != "System" {
		t.Errorf("SystemLabel: %q", tui.cfg.SystemLabel)
	}
	if tui.progress != -1 {
		t.Errorf("progress should be -1, got %f", tui.progress)
	}
}

func TestNewCustomTheme(t *testing.T) {
	tui := New(Config{Theme: ThemeBlue})
	if tui.theme != ThemeBlue {
		t.Error("expected ThemeBlue")
	}
}

func TestInputEnabled(t *testing.T) {
	tui := New(Config{})
	if !tui.inputEnabled() {
		t.Error("input should be enabled by default")
	}
	disabled := false
	tui2 := New(Config{InputEnabled: &disabled})
	if tui2.inputEnabled() {
		t.Error("input should be disabled")
	}
}

func TestSetThemeNilNoOp(t *testing.T) {
	tui := New(Config{Theme: ThemeBlue})
	tui.SetTheme(nil)
	if tui.theme != ThemeBlue {
		t.Error("SetTheme(nil) should be a no-op")
	}
}

func TestIsStreaming(t *testing.T) {
	tui := New(Config{})
	if tui.IsStreaming() {
		t.Error("should not be streaming initially")
	}
	tui.output.StartStreaming()
	if !tui.IsStreaming() {
		t.Error("should be streaming after StartStreaming")
	}
}

func TestStopStreamingIdempotent(t *testing.T) {
	tui := New(Config{})
	// StopStreaming when not streaming should not panic
	tui.StopStreaming()
}

func TestSetProgress(t *testing.T) {
	tui := New(Config{})
	tui.SetProgress("Loading", 0.5)
	if tui.progress != 0.5 {
		t.Errorf("progress: %f", tui.progress)
	}
	// clamp below 0
	tui.SetProgress("", -1)
	if tui.progress != 0 {
		t.Errorf("clamp below 0: %f", tui.progress)
	}
	// clamp above 1
	tui.SetProgress("", 2)
	if tui.progress != 1 {
		t.Errorf("clamp above 1: %f", tui.progress)
	}
	tui.ClearProgress()
	if tui.progress != -1 {
		t.Errorf("ClearProgress: %f", tui.progress)
	}
}

func TestExitSetsQuit(t *testing.T) {
	tui := New(Config{})
	tui.Exit()
	if !tui.quit {
		t.Error("Exit should set quit=true")
	}
}
